package toyorm

import (
	"database/sql"
	"os"
	"reflect"
)

type DBValSelector interface {
	Select(int) int
}

type DBPrimarySelector func(n int, key ...interface{}) int

func dbPrimaryKeySelector(n int, keys ...interface{}) int {
	sum := 0
	for _, k := range keys {
		switch val := k.(type) {
		case int:
			sum += val

		case int32:
			sum += int(val)
		case uint:
			sum += int(val)
		case uint32:
			sum += int(val)
		default:
			panic("primary key type not match")
		}
	}
	return sum % n
}

type ToyCollection struct {
	dbs                      []*sql.DB
	DefaultHandlerChain      map[string]CollectionHandlersChain
	DefaultModelHandlerChain map[*Model]map[string]CollectionHandlersChain
	ToyKernel
}

func OpenCollection(driverName string, dataSourceName ...string) (*ToyCollection, error) {
	t := ToyCollection{
		ToyKernel: ToyKernel{
			CacheModels:       map[reflect.Type]*Model{},
			CacheMiddleModels: map[reflect.Type]*Model{},
			belongToPreload:   map[*Model]map[string]*BelongToPreload{},
			oneToOnePreload:   map[*Model]map[string]*OneToOnePreload{},
			oneToManyPreload:  map[*Model]map[string]*OneToManyPreload{},
			// because have isRight feature need two point to save
			manyToManyPreload: map[*Model]map[string]map[bool]*ManyToManyPreload{},
			Logger:            os.Stdout,
		},
		DefaultHandlerChain: map[string]CollectionHandlersChain{
			"CreateTable":           {CollectionHandlerSimplePreload("CreateTable"), CollectionHandlerAssignToAllDb, CollectionHandlerCreateTable},
			"CreateTableIfNotExist": {CollectionHandlerSimplePreload("CreateTableIfNotExist"), CollectionHandlerAssignToAllDb, CollectionHandlerExistTableAbort, CollectionHandlerCreateTable},
			"DropTableIfExist":      {CollectionHandlerDropTablePreload("DropTableIfExist"), CollectionHandlerAssignToAllDb, CollectionHandlerNotExistTableAbort, CollectionHandlerDropTable},
			"DropTable":             {CollectionHandlerDropTablePreload("DropTable"), CollectionHandlerAssignToAllDb, CollectionHandlerDropTable},
			"Insert":                {CollectionHandlerPreloadContainerCheck, CollectionHandlerPreloadInsertOrSave("Insert"), CollectionHandlerInsertTimeGenerate, CollectionHandlerInsertAssignDbIndex, CollectionHandlerInsert},
			"Find":                  {CollectionHandlerPreloadContainerCheck, CollectionHandlerSoftDeleteCheck, CollectionHandlerPreloadFind, CollectionHandlerFindAssignDbIndex, CollectionHandlerFind},
			"FindOne":               {CollectionHandlerPreloadContainerCheck, CollectionHandlerSoftDeleteCheck, CollectionHandlerPreloadFind, CollectionHandlerFindOneAssignDbIndex, CollectionHandlerFindOne},
			"Update":                {CollectionHandlerSoftDeleteCheck, CollectionHandlerUpdateTimeGenerate, CollectionHandlerUpdateAssignDbIndex, CollectionHandlerUpdate},
			"Save":                  {CollectionHandlerPreloadContainerCheck, CollectionHandlerPreloadInsertOrSave("Save"), CollectionHandlerSaveTimeGenerate, CollectionHandlerInsertAssignDbIndex, CollectionHandlerSave},
			//"HardDelete":               {HandlerPreloadDelete, HandlerHardDelete},
			//"SoftDelete":               {HandlerPreloadDelete, HandlerSoftDelete},
			//"HardDeleteWithPrimaryKey": {HandlerPreloadDelete, HandlerSearchWithPrimaryKey, HandlerHardDelete},
			//"SoftDeleteWithPrimaryKey": {HandlerPreloadDelete, HandlerSearchWithPrimaryKey, HandlerSoftDelete},
		},
		DefaultModelHandlerChain: map[*Model]map[string]CollectionHandlersChain{},
	}
	switch driverName {
	case "mysql":
		t.Dialect = MySqlDialect{}
	case "sqlite3":
		t.Dialect = Sqlite3Dialect{}
	default:
		panic(ErrNotMatchDialect)
	}
	for _, source := range dataSourceName {
		db, err := sql.Open(driverName, source)
		if err != nil {
			t.Close()
			return nil, err
		}
		t.dbs = append(t.dbs, db)
	}
	return &t, nil
}

func (t *ToyCollection) Model(v interface{}) *CollectionBrick {
	var model *Model
	vType := LoopTypeIndirect(reflect.ValueOf(v).Type())
	// lazy init model
	model = t.GetModel(vType)
	brick := NewCollectionBrick(t, model)
	return brick
}

func (t *ToyCollection) ModelHandlers(option string, model *Model) CollectionHandlersChain {
	handlers := make(CollectionHandlersChain, 0, len(t.DefaultHandlerChain[option])+len(t.DefaultModelHandlerChain[model][option]))
	handlers = append(handlers, t.DefaultModelHandlerChain[model][option]...)
	handlers = append(handlers, t.DefaultHandlerChain[option]...)
	return handlers
}

func (t *ToyCollection) SetModelHandlers(option string, model *Model, handlers CollectionHandlersChain) {
	if t.DefaultModelHandlerChain[model] == nil {
		t.DefaultModelHandlerChain[model] = map[string]CollectionHandlersChain{}
	}
	t.DefaultModelHandlerChain[model][option] = handlers
}

func (t *ToyCollection) Close() error {
	errs := ErrCollectionQueryRow{}
	for i, db := range t.dbs {
		err := db.Close()
		if err != nil {
			errs[i] = err
		}
	}
	return errs
}
