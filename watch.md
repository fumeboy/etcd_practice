# Watch 

Etcd v3 的watch机制支持watch某个固定的key，也支持watch一个范围（可以用于模拟目录的结构的watch），所以 watchGroup 包含两种watcher，一种是 key watchers，数据结构是每个key对应一组watcher，另外一种是 range watchers, 数据结构是一个 IntervalTree，方便通过区间查找到对应的watcher。

每个 WatchableStore 包含两种 watcherGroup，一种是synced，一种是unsynced，前者表示该group的watcher数据都已经同步完毕，在等待新的变更，后者表示该group的watcher数据同步落后于当前最新变更，还在追赶。 当 Etcd 收到客户端的watch请求，如果请求携带了revision参数，则比较请求的revision和store当前的revision，如果大于当前revision，则放入synced组中，否则放入unsynced组。

同时 Etcd 会启动一个后台的goroutine持续同步unsynced的watcher，然后将其迁移到synced组。也就是这种机制下，Etcd v3 支持从任意版本开始watch，没有v2的1000条历史event表限制的问题（当然这是指没有compact的情况下）。 另外我们前面提到的，Etcd v2在通知客户端时，如果网络不好或者客户端读取比较慢，发生了阻塞，则会直接关闭当前连接，客户端需要重新发起请求。Etcd v3为了解决这个问题，专门维护了一个推送时阻塞的watcher队列，在另外的goroutine里进行重试。 Etcd v3 对过期机制也做了改进，过期时间设置在lease上，然后key和lease关联。这样可以实现多个key关联同一个lease id，方便设置统一的过期时间，以及实现批量续约。


## 接口定义
WatchService 只有一个 `Watch` 方法。

```java
// Watch 观察将要发生或者已经发生的事件。
// 输入和输出都是流;输入流用于创建和取消观察，而输出流发送事件。
// 一个观察 RPC 可以在一次性在多个 key 范围上观察，并为多个观察流化事件。
// 整个事件历史可以从最后压缩修订版本开始观察。
rpc Watch(stream WatchRequest) returns (stream WatchResponse) {}
```

## 消息定义

请求的消息体是 `WatchRequest`：

```java
message WatchRequest {
  // request_union 要么是创建新的观察者的请求，要么是取消一个已经存在的观察者的请求
  oneof request_union {
    WatchCreateRequest create_request = 1;
    WatchCancelRequest cancel_request = 2;
  }
}
```

创建新的观察者的请求 `WatchCreateRequest`：

```java
message WatchCreateRequest {
  // key 是注册要观察的 key
  bytes key = 1;

  // range_end 是要观察的范围 [key, range_end) 的终点。
  // 如果 range_end 没有设置，则只有参数 key 被观察。
  // 如果 range_end 等同于 '\0'， 则大于等于参数 key 的所有 key 都将被观察
  // 如果 range_end 比给定 key 大1， 则所有以给定 key 为前缀的 key 都将被观察
  bytes range_end = 2;

  // start_revision 是可选的开始(包括)观察的修订版本。不设置 start_revision 则表示 "现在".
  int64 start_revision = 3;

  // 设置 progress_notify ，这样如果最近没有事件，etcd 服务器将定期的发送不带任何事件的 WatchResponse 给新的观察者。
  // 当客户端希望从最近已知的修订版本开始恢复断开的观察者时有用。
  // etcd 服务器将基于当前负载决定它发送通知的频率。
  bool progress_notify = 4;

  enum FilterType {
  // 过滤掉 put 事件
  NOPUT = 0;

  // 过滤掉 delete 事件
  NODELETE = 1;
  }

  // 过滤器，在服务器端发送事件给回观察者之前，过滤掉事件。
  repeated FilterType filters = 5;

  // 如果 prev_kv 被设置，被创建的观察者在事件发生前获取上一次的KV。
  // 如果上一次的KV已经被压缩，则不会返回任何东西
  bool prev_kv = 6;
}
```

取消已有观察者的 `WatchCancelRequest` ：

```java
message WatchCancelRequest {
  // watch_id 是要取消的观察者的id，这样就不再有更多事件传播过来了。
  int64 watch_id = 1;
}
```

应答的消息体 `WatchResponse`：

```java
message WatchResponse {
  ResponseHeader header = 1;
  // watch_id 是和应答相关的观察者的ID
  int64 watch_id = 2;

  // 如果应答是用于创建观察者请求的，则 created 设置为 true。
  // 客户端应该记录 watch_id 并期待从同样的流中为创建的观察者接收事件。
  // 所有发送给被创建的观察者的事件将附带同样的 watch_id
  bool created = 3;

  // 如果应答是用于取消观察者请求的，则 canceled 设置为true。
  // 不会再有事件发送给被取消的观察者。
  bool canceled = 4;

  // compact_revision 被设置为最小 index，如果观察者试图观察被压缩的 index。
  // 当在被压缩的修订版本上创建观察者或者观察者无法追上键值对存储的进展时发生。
  // 客户端应该视观察者为被取消，并不应该试图再次创建任何带有相同 start_revision 的观察者。
  int64 compact_revision  = 5;

  // cancel_reason 指出取消观察者的理由.
  string cancel_reason = 6;

  repeated mvccpb.Event events = 11;
}
```

mvccpb.Event 的消息体：

```java
message Event {
  enum EventType {
    PUT = 0;
    DELETE = 1;
  }

  // type 是事件的类型。
  // 如果类型是 PUT，表明新的数据已经存储到 key。
  // 如果类型是 DELETE， 表明 key 已经被删除。
  EventType type = 1;

  // kv 为事件持有 KeyValue。
  // PUT 事件包含当前的kv键值对
  // kv.Version=1 的 PUT 事件表明 key 的创建
  // DELETE/EXPIRE 事件包含被删除的 key，它的修改修订版本设置为删除的修订版本
  KeyValue kv = 2;

  // prev_kv 持有在事件发生前的键值对
  KeyValue prev_kv = 3;
}
```

