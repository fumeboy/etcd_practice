        
# proto 定义

```proto
service KV {
  // 从键值存储中获取范围内的key.
  rpc Range(RangeRequest) returns (RangeResponse) {}

  // 放置给定key到键值存储.
  // put请求增加键值存储的修订版本并在事件历史中生成一个事件.
  rpc Put(PutRequest) returns (PutResponse) {}

  // 从键值存储中删除给定范围。
  // 删除请求增加键值存储的修订版本并在事件历史中为每个被删除的key生成一个删除事件.
  rpc DeleteRange(DeleteRangeRequest) returns (DeleteRangeResponse) {}

  // 在单个事务中处理多个请求。
  // 一个 txn 请求增加键值存储的修订版本并为每个完成的请求生成带有相同修订版本的事件。
  // 不容许在一个txn中多次修改同一个key.
  rpc Txn(TxnRequest) returns (TxnResponse) {}

  // 压缩在etcd键值存储中的事件历史。
  // 键值存储应该定期压缩，否则事件历史会无限制的持续增长.
  rpc Compact(CompactionRequest) returns (CompactionResponse) {}
}
```


# Range 方法

Range方法从键值存储中获取范围内的key.

```java
rpc Range(RangeRequest) returns (RangeResponse) {}
```

> 注意: 没有操作单个key的方法，即使是存取单个key，也是需要使用 `Range` 方法的。

### use
`etcdctl get foo`

`curl -L http://100.10.0.2:2379/v3/kv/range  -X POST -d '{"key": "Zm9v"}'`

#### important
http 方式交互时，key 和 value 的值都是 base64 字符串

foo is 'Zm9v' in Base64 and bar is 'YmFy'

### 请求体 RangeRequest

```java
message RangeRequest {
  enum SortOrder {
	NONE = 0; // 默认, 不排序
	ASCEND = 1; // 正序，低的值在前
	DESCEND = 2; // 倒序，高的值在前
  }
  enum SortTarget {
	KEY = 0;
	VERSION = 1;
	CREATE = 2;
	MOD = 3;
	VALUE = 4;
  }

  // key是 range 的第一个 key。如果 range_end 没有指定，请求仅查找这个key
  bytes key = 1;

  // range_end 是请求范围的上限 [key, range_end)
  // 如果 range_end 是 '\0'，范围是大于等于 key 的所有 key。
  // 如果 range_end 是 key 加一(例如, "aa"+1 == "ab", "a\xff"+1 == "b")， 那么 range 请求获取以 key 为前缀的所有 key
  // 如果 key 和 range_end 都是'\0'，则 range 查询返回所有 key
  bytes range_end = 2;

  // 请求返回的 key 的数量限制。如果 limit 设置为0，则视为没有限制
  int64 limit = 3;

  // 修订版本是用于 range 的键值对存储的时间点。
  // 如果修订版本小于或等于零，range 是用在最新的键值对存储上。
  // 如果指定修订版本已经被压缩，返回 ErrCompacted 作为应答
  int64 revision = 4;

  // 指定返回结果的排序顺序
  SortOrder sort_order = 5;

  // 用于排序的键值字段
  SortTarget sort_target = 6;

  // 设置 range 请求使用串行化成员本地读(serializable member-local read)。
  // range 请求默认是线性化的;线性化请求相比串行化请求有更高的延迟和低吞吐量，但是反映集群当前的一致性。
  // 为了更好的性能，以可能脏读为交换，串行化范围请求在本地处理，无需和集群中的其他节点达到一致。
  bool serializable = 7;

  // keys_only 被设置时仅返回 key 而不需要 value
  bool keys_only = 8;

  // count_only 被设置时仅仅返回范围内 key 的数量
  bool count_only = 9;

  // min_mod_revision 是返回 key 的 mod revision 的下限；更低 mod revision 的所有 key 都将被过滤掉
  int64 min_mod_revision = 10;

  // max_mod_revision 是返回 key 的 mod revision 的上限；更高 mod revision 的所有 key 都将被过滤掉
  int64 max_mod_revision = 11;

    // min_create_revision 是返回 key 的 create revision 的下限；更低 create revision 的所有 key 都将被过滤掉
  int64 min_create_revision = 12;

  // max_create_revision 是返回 key 的 create revision 的上限；更高 create revision 的所有 key 都将被过滤掉
  int64 max_create_revision = 13;
}
```

### 应答体 RangeResponse

```java
message RangeResponse {
  ResponseHeader header = 1;

  // kvs 是匹配 range 请求的键值对列表
  // 当 count 时是空的
  repeated mvccpb.KeyValue kvs = 2;

  // more 代表在被请求的范围内是否还有更多的 key
  bool more = 3;

  // count 被设置为在范围内的 key 的数量
  int64 count = 4;
}
```

`mvccpb.KeyValue` 来自 `kv.proto`，消息体定义为：

```java
message KeyValue {
  // key 是 bytes 格式的 key。不容许 key 为空。
  bytes key = 1;

  // create_revision 是这个 key 最后一次创建的修订版本
  int64 create_revision = 2;

  // mod_revision 是这个 key 最后一次修改的修订版本
  int64 mod_revision = 3;

  // version 是 key 的版本。删除会重置版本为0,而任何 key 的修改会增加它的版本。
  int64 version = 4;

  // value 是 key 持有的值，bytes 格式。
  bytes value = 5;

  // lease 是附加给 key 的租约 id。
  // 当附加的租约过期时，key 将被删除。
  // 如果 lease 为0,则没有租约附加到 key。
  int64 lease = 6;
}
```

# Put 方法

Put 方法设置指定 key 到键值存储.

Put 方法增加键值存储的修订版本并在事件历史中生成一个事件.

```java
rpc Put(PutRequest) returns (PutResponse) {}
```

### use
`etcdctl put foo bar`

`curl -L http://100.10.0.2:2379/v3/kv/put -X POST -d '{"key": "Zm9v", "value": "YmFy"}'`

### PutRequest

```java
message PutRequest {
  // byte 数组形式的 key，用来保存到键值对存储
  bytes key = 1;

  // byte 数组形式的 value，在键值对存储中和 key 关联
  bytes value = 2;

  // 在键值存储中和 key 关联的租约id。0代表没有租约。
  int64 lease = 3;

  // 如果 prev_kv 被设置，etcd 获取改变之前的上一个键值对。
  // 上一个键值对将在 put 应答中被返回
  bool prev_kv = 4;

  // 如果 ignore_value 被设置, etcd 使用它当前的 value 更新 key.
  // 如果 key 不存在，返回错误.
  bool ignore_value = 5;

  // 如果 ignore_lease 被设置, etcd 使用它当前的租约更新 key.
  // 如果 key 不存在，返回错误.
  bool ignore_lease = 6;
}
```

### PutResponse

```java
message PutResponse {
  ResponseHeader header = 1;

  // 如果请求中的 prev_kv 被设置，将会返回上一个键值对
  mvccpb.KeyValue prev_kv = 2;
}
```

# DeleteRange 方法

DeleteRange 方法从键值存储中删除给定范围。

删除请求增加键值存储的修订版本,并在事件历史中为每个被删除的 key 生成一个删除事件.

```java
rpc DeleteRange(DeleteRangeRequest) returns (DeleteRangeResponse)
```

### use

`etcdctl del foo`

### DeleteRangeRequest

```java
message DeleteRangeRequest {
  // key是要删除的范围的第一个key
  bytes key = 1;

  // range_end 是要删除范围[key, range_end)的最后一个key
  // 如果 range_end 没有给定，范围定义为仅包含 key 参数
  // 如果 range_end 比给定的 key 大1，则 range 是以给定 key 为前缀的所有 key
  // 如果 range_end 是 '\0'， 范围是所有大于等于参数 key 的所有 key。
  bytes range_end = 2;

  // 如果 prev_kv 被设置，etcd获取删除之前的上一个键值对。
  // 上一个键值对将在 delete 应答中被返回
  bool prev_kv = 3;
}
```

### DeleteRangeResponse

```java
message DeleteRangeResponse {
  ResponseHeader header = 1;

  // 被范围删除请求删除的 key 的数量
  int64 deleted = 2;

  // 如果请求中的 prev_kv 被设置，将会返回上一个键值对
  repeated mvccpb.KeyValue prev_kvs = 3;
}
```

# Txn 方法

Txn 方法在单个事务中处理多个请求。

txn 请求增加键值存储的修订版本并为每个完成的请求生成带有相同修订版本的事件。

不容许在一个 txn 中多次修改同一个 key。

```java
rpc Txn(TxnRequest) returns (TxnResponse) {}
```

### 背景

以下内容翻译来自 proto文件中 TxnRequest 的注释，解释了Txn请求的工作方式.

> 来自 google paxosdb 论文:
>
> 我们的实现围绕强大的我们称为 `MultiOp` 的原生(primitive)。除了游历外的所有其他数据库操作被实现为对 MultiOp 的单一调用。MultiOp 被原子执行并由三个部分组成：
>
> 1. 被称为 `guard` 的测试列表。在 guard 中每个测试检查数据库中的单个项。它可能检查某个值的存在或者缺失，或者和给定的值比较。在 guard 中两个不同的测试可能应用于数据库中相同或者不同的项。guard 中的所有测试被应用然后 MultiOp 返回结果。如果所有测试是 true，MultiOp 执行 t 操作 (见下面的第二项), 否则它执行 f 操作 (见下面的第三项).
> 2. 被称为 `t` 操作的数据库操作列表. 列表中的每个操作是插入，删除，或者查找操作，并应用到单个数据库项。列表中的两个不同操作可能应用到数据库中相同或者不同的项。如果 guard 评价为true 这些操作将被执行
> 3. 被成为 f 操作的数据库操作列表. 类似 t 操作, 但是是在 guard 评价为 false 时执行。

### 请求体 TxnRequest

```java
message TxnRequest {
  // compare 是断言列表，体现为条件的联合。
  // 如果比较成功，那么成功请求将被按顺序处理，而应答将按顺序包含他们对应的应答。
  // 如果比较失败，那么失败请求将被按顺序处理，而应答将按顺序包含他们对应的应答。
  repeated Compare compare = 1;

  // 成功请求列表，当比较评估为 true 时将被应用。
  repeated RequestOp success = 2;

  // 失败请求列表，当比较评估为 false 时将被应用。
  repeated RequestOp failure = 3;
}
```

#### Compare

```java
message Compare {
  enum CompareResult {
    EQUAL = 0;
    GREATER = 1;
    LESS = 2;
    NOT_EQUAL = 3;
  }
  enum CompareTarget {
    VERSION = 0;
    CREATE = 1;
    MOD = 2;
    VALUE= 3;
  }

  // result 是这个比较的逻辑比较操作
  CompareResult result = 1;

  // target 是比较要检查的键值字段
  CompareTarget target = 2;

  // key 是用于比较操作的主题key
  bytes key = 3;

  oneof target_union {
    // version 是给定 key 的版本
    int64 version = 4;

    // create_revision 是给定 key 的创建修订版本
    int64 create_revision = 5;

    // mod_revision 是给定 key 的最后修改修订版本
    int64 mod_revision = 6;

    // value 是给定 key 的值，以 bytes 的形式
    bytes value = 7;
  }
}
```

#### RequestOp

```java
message RequestOp {
  // request 是可以被事务接受的请求类型的联合
  oneof request {
    RangeRequest request_range = 1;
    PutRequest request_put = 2;
    DeleteRangeRequest request_delete_range = 3;
  }
}

```

### 响应体 TxnResponse

```java
message TxnResponse {
  ResponseHeader header = 1;

  // 如果比较评估为true则succeeded被设置为true，否则是false
  bool succeeded = 2;

  // 应答列表，如果 succeeded 是 true 则对应成功请求，如果 succeeded 是 false 则对应失败请求
  repeated ResponseOp responses = 3;
}
```

#### ResponseOp

```java
message ResponseOp {
  // response 是事务返回的应答类型的联合
  oneof response {
    RangeResponse response_range = 1;
    PutResponse response_put = 2;
    DeleteRangeResponse response_delete_range = 3;
  }
}
```

# Compact 方法

Compact 方法压缩 etcd 的事件历史。

应该定期压缩，否则事件历史会无限制的持续增长.

```java
rpc Compact(CompactionRequest) returns (CompactionResponse) {}
```

### CompactionRequest

CompactionRequest 压缩键值对存储到给定修订版本。所有修订版本比压缩修订版本小的键都将被删除：

```java
message CompactionRequest {
  // 键值存储的修订版本，用于比较操作
  int64 revision = 1;

  // physical设置为 true 时 RPC 将会等待直到压缩物理性的应用到本地数据库，到这程度被压缩的项将完全从后端数据库中移除。
  bool physical = 2;
}
```

### CompactionResponse

```java
message CompactionResponse {
  ResponseHeader header = 1;
}
```

## 版本机制

Etcd v3 store 分为两部分，一部分是内存中的索引，kvindex，是基于google开源的一个golang的btree实现的，另外一部分是后端存储

按照它的设计，backend可以对接多种存储，当前使用的boltdb。boltdb是一个单机的支持事务的kv存储，Etcd 的事务是基于boltdb的事务实现的。Etcd 在boltdb中存储的key是revision，value是 Etcd 自己的key-value组合，也就是说 Etcd 会在boltdb中把每个版本都保存下，从而实现了多版本机制。

举例来说：

用etcdctl通过批量接口写入两条记录：

```shell script
etcdctl txn <<<'
put key1 "v1"
put key2 "v2"

'
```

再通过批量接口更新这两条记录：

```shell script
etcdctl txn <<<'
put key1 "v12"
put key2 "v22"
'
```

boltdb中其实有了4条数据：

```text
rev={3 0}, key=key1, value="v1"
rev={3 1}, key=key2, value="v2"
rev={4 0}, key=key1, value="v12"
rev={4 1}, key=key2, value="v22"
```

revision 由两部分组成，第一部分main rev，每次事务进行加一，第二部分sub rev，同一个事务中的每次操作加一。

如上示例，第一次操作的main rev是3，第二次是4。当然这种机制大家想到的第一个问题就是空间问题，所以 Etcd 提供了命令和设置选项来控制compact，同时支持put操作的参数来精确控制某个key的历史版本数。 

了解了 Etcd 的磁盘存储，可以看出如果要从boltdb中查询数据，必须通过revision，但客户端都是通过key来查询value，所以 Etcd 的内存kvindex保存的就是key和revision之前的映射关系，用来加速查询。