package toyorm

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
)

type DBValSelector interface {
	Select(int) int
}

type DBPrimarySelector func(key interface{}, n int) int

type ToyCollection struct {
	dbs                      []*sql.DB
	DefaultDBSelector        map[string]DBPrimarySelector
	DefaultHandlerChain      map[string]CollectionHandlersChain
	DefaultModelHandlerChain map[string]map[*Model]CollectionHandlersChain
	DBSelect                 func(v interface{}) int
	ToyKernel
}

func (t *ToyCollection) Close() error {
	var errs []error
	for _, db := range t.dbs {
		if err := db.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.New(fmt.Sprintf("%v\n", errs))
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
			"CreateTable":           {CollectionHandlerSimplePreload("CreateTable"), CollectionHandlerCreateTable},
			"CreateTableIfNotExist": {CollectionHandlerSimplePreload("CreateTableIfNotExist"), CollectionHandlerExistTableAbort, CollectionHandlerCreateTable},
			"DropTableIfExist":      {CollectionHandlerDropTablePreload("DropTableIfExist"), CollectionHandlerNotExistTableAbort, CollectionHandlerDropTable},
			"DropTable":             {CollectionHandlerDropTablePreload("DropTable"), CollectionHandlerDropTable},
			"Insert":                {CollectionHandlerPreloadContainerCheck, CollectionHandlerPreloadInsertOrSave("Insert"), CollectionHandlerInsertTimeGenerate, CollectionHandlerInsert},
			//"Find":                     {HandlerPreloadContainerCheck, HandlerSoftDeleteCheck, HandlerFind, HandlerPreloadFind},
			//"Update":                   {HandlerSoftDeleteCheck, HandlerUpdateTimeGenerate, HandlerUpdate},
			"Save": {CollectionHandlerPreloadContainerCheck, CollectionHandlerPreloadInsertOrSave("Save"), CollectionHandlerSaveTimeGenerate, CollectionHandlerSave},
			//"HardDelete":               {HandlerPreloadDelete, HandlerHardDelete},
			//"SoftDelete":               {HandlerPreloadDelete, HandlerSoftDelete},
			//"HardDeleteWithPrimaryKey": {HandlerPreloadDelete, HandlerSearchWithPrimaryKey, HandlerHardDelete},
			//"SoftDeleteWithPrimaryKey": {HandlerPreloadDelete, HandlerSearchWithPrimaryKey, HandlerSoftDelete},
		},
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

func (t *ToyCollection) Model(v interface{}, selector DBPrimarySelector) *CollectionBrick {
	var model *Model
	vType := LoopTypeIndirect(reflect.ValueOf(v).Type())
	// lazy init model
	model = t.GetModel(vType)
	t.DefaultDBSelector[model.Name] = selector
	brick := NewCollectionBrick(t, model)
	return brick
}

func (t *ToyCollection) ModelHandlers(option string, model *Model) CollectionHandlersChain {
	handlers := make(CollectionHandlersChain, 0, len(t.DefaultHandlerChain[option])+len(t.DefaultModelHandlerChain[option][model]))
	handlers = append(handlers, t.DefaultModelHandlerChain[option][model]...)
	handlers = append(handlers, t.DefaultHandlerChain[option]...)
	return handlers
}
