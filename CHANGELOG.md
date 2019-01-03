- [FIX] order by conflict with group by

### Toyorm v0.6.0-alpha (Dec 28 2018)
- [NEW] add MustAddr method for Toy
- [CHANGE] In Insert operation, CreatedAt field will keep the value unchanged when it is not zero
- [FIX] invoke IsZero with not struct value will panic
- [NEW] add default tag match default value in database
- [CHANGE] brick.Template method , the insertion part has changed

### Toyorm v0.5.3-alpha (Dec 17 2018)
- [FIX] Count does not work when there is Join statement
- [FIX] USave will generate new createdAt value

### Toyorm v0.5.2-alpha (Dec 17 2018)
- [FIX] get custom default primary key not work
- [FIX] allow insert/save with zero primary key
- [FIX] order by conflict with limit/offset

### Toyorm v0.5.1-alpha (Dec 10 2018)
- [FIX] custom preload table name has dirty cache data

### Toyorm v0.5.0-alpha (Dec 9 2018)
- [NEW] custom table name

### Toyorm v0.4.2-alpha (Dec 2 2018)
- [NEW] go module support
- [FIX] save/insert with other struct type will nil field value

### Toyorm v0.4.1-alpha (Oct 11 2018)

- [FIX] many to many preload loss middle model log
- [FIX] many to many query loss related field data when primary key field type different from model primary key type

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
