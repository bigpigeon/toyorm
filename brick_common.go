/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

type BrickCommon struct {
	Model             *Model
	BelongToPreload   map[string]*BelongToPreload
	OneToOnePreload   map[string]*OneToOnePreload
	OneToManyPreload  map[string]*OneToManyPreload
	ManyToManyPreload map[string]*ManyToManyPreload

	// use by find/scan/insert/update/replace/where ,if FieldsSelector[Mode] is set,then ignoreModeSelector will failure
	// FieldsSelector[ModeDefault] can work on all Mode when they specified Mode not set
	FieldsSelector [ModeEnd][]Field
	// use by insert/update/replace/where  when source value is struct
	// ignoreMode IgnoreMode
	ignoreModeSelector [ModeEnd]IgnoreMode
}

func (t *BrickCommon) CopyBelongToPreload() map[string]*BelongToPreload {
	preloadMap := map[string]*BelongToPreload{}
	for k, v := range t.BelongToPreload {
		preloadMap[k] = v
	}
	return preloadMap
}

func (t *BrickCommon) CopyOneToOnePreload() map[string]*OneToOnePreload {
	preloadMap := map[string]*OneToOnePreload{}
	for k, v := range t.OneToOnePreload {
		preloadMap[k] = v
	}
	return preloadMap
}

func (t *BrickCommon) CopyOneToManyPreload() map[string]*OneToManyPreload {
	preloadMap := map[string]*OneToManyPreload{}
	for k, v := range t.OneToManyPreload {
		preloadMap[k] = v
	}
	return preloadMap
}

func (t *BrickCommon) CopyManyToManyPreload() map[string]*ManyToManyPreload {
	preloadMap := map[string]*ManyToManyPreload{}
	for k, v := range t.ManyToManyPreload {
		preloadMap[k] = v
	}
	return preloadMap
}

func (t *BrickCommon) getFieldValuePairWithRecord(mode Mode, record ModelRecord) []ColumnValue {
	var fields []Field
	if len(t.FieldsSelector[mode]) > 0 {
		fields = t.FieldsSelector[mode]
	} else if len(t.FieldsSelector[ModeDefault]) > 0 {
		fields = t.FieldsSelector[ModeDefault]
	}

	var useIgnoreMode bool
	if len(fields) == 0 {
		fields = t.Model.GetSqlFields()
		useIgnoreMode = record.IsVariableContainer() == false
	}
	var columnValues []ColumnValue
	if useIgnoreMode {
		for _, mField := range fields {
			if fieldValue := record.Field(mField.Name()); fieldValue.IsValid() {
				if t.ignoreModeSelector[mode].Ignore(fieldValue) == false {
					if mField.IsPrimary() && IsZero(fieldValue) {

					} else {
						columnValues = append(columnValues, &modelFieldValue{mField, fieldValue})
					}
				}
			}
		}
	} else {
		for _, mField := range fields {
			if fieldValue := record.Field(mField.Name()); fieldValue.IsValid() {
				columnValues = append(columnValues, &modelFieldValue{mField, fieldValue})
			}
		}
	}
	return columnValues
}

func (t *BrickCommon) getSelectFields(records ModelRecordFieldTypes) []Column {
	var fields []Field
	if len(t.FieldsSelector[ModeSelect]) > 0 {
		fields = t.FieldsSelector[ModeSelect]
	} else if len(t.FieldsSelector[ModeDefault]) > 0 {
		fields = t.FieldsSelector[ModeDefault]
	} else {
		fields = t.Model.GetSqlFields()
	}
	fields = getFieldsWithRecords(fields, records)
	var columns []Column
	for _, field := range fields {
		columns = append(columns, field)
	}
	return columns
}

func (t *BrickCommon) getScanFields(records ModelRecordFieldTypes) []Field {
	var fields []Field
	if len(t.FieldsSelector[ModeScan]) > 0 {
		fields = t.FieldsSelector[ModeScan]
	} else if len(t.FieldsSelector[ModeDefault]) > 0 {
		fields = t.FieldsSelector[ModeDefault]
	} else {
		fields = t.Model.GetSqlFields()
	}
	return getFieldsWithRecords(fields, records)
}

// use for order by
func (t *BrickCommon) ToDesc(v interface{}) Column {
	field := t.Model.fieldSelect(v)

	column := StrColumn(field.Column() + " DESC")
	return column
}
