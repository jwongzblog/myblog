从mongodb [v3.2](https://docs.mongodb.com/manual/tutorial/remove-documents/)版本开始，删除对象提供了一对新接口：deleteOne/deleteMany，也是官方推荐使用并取代原先的remove接口，remove接口仍然保留

# 原因
remove 相比于deleteOne/deleteMany接口多了一次justOne的判断，可以节省一个cpu的时钟

# 源码分析
```
curd_api.js

        } else if (op.deleteOne) {
            if (!op.deleteOne.filter) {
                throw new Error('deleteOne bulkWrite operation expects the filter field');
            }

            // Translate operation to bulkOp operation.
            var deleteOp = bulkOp.find(op.deleteOne.filter);

            if (op.deleteOne.collation) {
                deleteOp.collation(op.deleteOne.collation);
            }

            deleteOp.removeOne();
        } else if (op.deleteMany) {
            if (!op.deleteMany.filter) {
                throw new Error('deleteMany bulkWrite operation expects the filter field');
            }

            // Translate operation to bulkOp operation.
            var deleteOp = bulkOp.find(op.deleteMany.filter);

            if (op.deleteMany.collation) {
                deleteOp.collation(op.deleteMany.collation);
            }

            deleteOp.remove();
        }
```

```
collection.js

DBCollection.prototype.remove = function(t, justOne) {
    var parsed = this._parseRemove(t, justOne);
    var query = parsed.query;
    var justOne = parsed.justOne;
    var wc = parsed.wc;
    var collation = parsed.collation;

    var result = undefined;
    var startTime =
        (typeof(_verboseShell) === 'undefined' || !_verboseShell) ? 0 : new Date().getTime();

    if (this.getMongo().writeMode() != "legacy") {
        var bulk = this.initializeOrderedBulkOp();
        var removeOp = bulk.find(query);

        if (collation) {
            removeOp.collation(collation);
        }

        if (justOne) {
            removeOp.removeOne();
        } else {
            removeOp.remove();
        }
```

# others
- 所有的删除接口都是先find object，再执行删除，这也是为什么批量删除时，cache页置换速率翻了好几倍的原因