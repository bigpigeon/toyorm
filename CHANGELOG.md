### Toyorm v0.4-alpha (Jul 2 2018)

- [New] toy-factory ,use sql table to create Model's bind struct
- [NEW] insert/update/find benchmark
- [NEW] USave operation, use Update operation to save data
- [NEW] Cas special Model Field(use in Save/USave)
- [FIX] debugPrint args format
- [REMOVE] BrickColumn and it's related interface
- [REMOVE] CreatedAt search in Save handlers


### Toyorm v0.3.2-alpha (May 16 2018)

- [New] toy-doctor https://github.com/bigpigeon/toy-doctor
- [New] website https://bigpigeon.org/toyorm

### Toyorm v0.3.1-alpha (Apr 01, 2018)

- [NEW] support Preload On Join
- [NEW] Toy/ToyCollection can SetDebug

### Toyorm v0.3-alpha (Apr 01, 2018)

- [NEW] support Join operation
- [NEW] ModelField tag:alias
- [NEW] ModelField tag:join

### Toyorm v0.2-alpha (Apr 01, 2018)

- [NEW] support CreateTable/DropTable/HasTable operation
- [NEW] support Insert/Save/Update/Find/Replace/Count sql operation
- [NEW] support Begin/Commit/Rollback transaction operation
- [NEW] support Where/Or/And/Limit/Offset/OrderBy/GroupBy Conditions operation
- [NEW] use result view sql error and action
- [NEW] use ToyBrick.Preload to association query/exec
- [NEW] use Template to custom sql exec/query
- [NEW] support mysql,sqlite3,postgresql database
- [NEW] database collection
- [NEW] support multiple selector
- [NEW] use ToyBrick.IgnoreMode to limit zero value
