/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

type PreloadType int

// this is describe one to one relationship with table and its sub table
// e.g select * from sub_table where id = (table.RelationField).value
type BelongToPreload struct {
	Model         *Model
	SubModel      *Model
	RelationField Field
	// used to save other table value
	ContainerField Field
}

// this is describe one to one relationship with table and its sub table
// e.g select * from sub_table where sub_table.RelationField = table.id;
type OneToOnePreload struct {
	Model         *Model
	SubModel      *Model
	RelationField Field
	// used to save other table value
	ContainerField Field
}

// this is describe one to many relationship with table and its sub table
//e.g select * from sub_table where sub_table.RelationField = table.id
type OneToManyPreload struct {
	Model          *Model
	SubModel       *Model
	RelationField  Field
	ContainerField Field
}

// this is describe many to many relationship with table and its sub table
// e.g select * from middle_table where table.id = table.id and sub_table.id = (table.RelationField).value
type ManyToManyPreload struct {
	MiddleModel      *Model
	Model            *Model
	SubModel         *Model
	ContainerField   Field
	RelationField    Field
	SubRelationField Field
}
