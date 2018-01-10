package toyorm

type PreloadType int

// this is describe one to one relationship with table and its sub table
// IsBelongTo means RelationField at table or sub_table;
// e.g
// if IsBelongTo is false=>select * from sub_table where sub_table.RelationField = table.id;
// if IsBelongTo is true=>select * from sub_table where id = (table.RelationField).value
//
type OneToOnePreload struct {
	IsBelongTo    bool
	Model         *Model
	SubModel      *Model
	RelationField *ModelField
	// used to save other table value
	ContainerField *ModelField
}

// this is describe one to many relationship with table and its sub table
//e.g select * from sub_table where sub_table.RelationField = table.id
type OneToManyPreload struct {
	Model          *Model
	SubModel       *Model
	RelationField  *ModelField
	ContainerField *ModelField
}

// this is describe many to many relationship with table and its sub table
// e.g select * from middle_table where table.id = table.id and sub_table.id = (table.RelationField).value
type ManyToManyPreload struct {
	MiddleModel    *Model
	Model          *Model
	SubModel       *Model
	ContainerField *ModelField
	IsRight        bool
}
