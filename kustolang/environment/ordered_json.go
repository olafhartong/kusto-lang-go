package environment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type orderedJSONField struct {
	Name  string
	Value json.RawMessage
}

func (e *EnvironmentDefinition) UnmarshalJSON(data []byte) error {
	type plain EnvironmentDefinition
	var decoded plain
	if err := decodeStrictJSON(data, &decoded); err != nil {
		return err
	}
	*e = EnvironmentDefinition(decoded)
	e.TableOrder, e.TableColumnOrder = orderedTableMetadata(data)
	e.ExternalTableOrder, e.ExternalTableColumnOrder = orderedTableMetadataForField(data, "external_tables")
	e.MaterializedViewOrder, e.MaterializedViewColumnOrder = orderedTableMetadataForField(data, "materialized_views")
	e.StoredQueryResultOrder, e.StoredQueryResultColumnOrder = orderedTableMetadataForField(data, "stored_query_results")
	return nil
}

func (l *LocalData) UnmarshalJSON(data []byte) error {
	type plain LocalData
	var decoded plain
	if err := decodeStrictJSON(data, &decoded); err != nil {
		return err
	}
	*l = LocalData(decoded)
	l.TableOrder, l.TableColumnOrder = orderedTableMetadata(data)
	l.ExternalTableOrder, l.ExternalTableColumnOrder = orderedTableMetadataForField(data, "external_tables")
	l.MaterializedViewOrder, l.MaterializedViewColumnOrder = orderedTableMetadataForField(data, "materialized_views")
	l.StoredQueryResultOrder, l.StoredQueryResultColumnOrder = orderedTableMetadataForField(data, "stored_query_results")
	return nil
}

func (d *DatabaseDetails) UnmarshalJSON(data []byte) error {
	type plain DatabaseDetails
	var decoded plain
	if err := decodeStrictJSON(data, &decoded); err != nil {
		return err
	}
	*d = DatabaseDetails(decoded)
	d.TableOrder, d.TableColumnOrder = orderedTableMetadata(data)
	d.ExternalTableOrder, d.ExternalTableColumnOrder = orderedTableMetadataForField(data, "external_tables")
	d.MaterializedViewOrder, d.MaterializedViewColumnOrder = orderedTableMetadataForField(data, "materialized_views")
	d.StoredQueryResultOrder, d.StoredQueryResultColumnOrder = orderedTableMetadataForField(data, "stored_query_results")
	return nil
}

func (d *ColumnTypeDetails) UnmarshalJSON(data []byte) error {
	type plain ColumnTypeDetails
	var decoded plain
	if err := decodeStrictJSON(data, &decoded); err != nil {
		return err
	}
	*d = ColumnTypeDetails(decoded)
	fields, err := decodeOrderedJSONObject(data)
	if err != nil {
		return err
	}
	seenFields := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		if _, duplicate := seenFields[field.Name]; duplicate {
			return fmt.Errorf("duplicate column type field %q", field.Name)
		}
		seenFields[field.Name] = struct{}{}
		if field.Name != "properties" {
			continue
		}
		properties, err := decodeOrderedJSONObject(field.Value)
		if err != nil {
			return err
		}
		d.PropertyOrder = make([]string, 0, len(properties))
		seen := make(map[string]struct{}, len(properties))
		for _, property := range properties {
			if _, duplicate := seen[property.Name]; duplicate {
				return fmt.Errorf("duplicate dynamic property %q", property.Name)
			}
			seen[property.Name] = struct{}{}
			d.PropertyOrder = append(d.PropertyOrder, property.Name)
		}
	}
	return nil
}

func (d ColumnTypeDetails) MarshalJSON() ([]byte, error) {
	return marshalColumnTypeDetails(d, nil, map[*ColumnTypeDetails]bool{})
}

func marshalColumnTypeDetails(d ColumnTypeDetails, source *ColumnTypeDetails, stack map[*ColumnTypeDetails]bool) ([]byte, error) {
	if source != nil {
		if stack[source] {
			return nil, fmt.Errorf("json: unsupported value: encountered a cycle via *analyzer.ColumnTypeDetails")
		}
		stack[source] = true
		defer delete(stack, source)
	}
	encodedType, err := json.Marshal(d.Type)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	buffer.WriteString(`{"type":`)
	buffer.Write(encodedType)
	if len(d.Properties) > 0 {
		buffer.WriteString(`,"properties":{`)
		for index, name := range orderedColumnTypeDetailsNames(d) {
			if index > 0 {
				buffer.WriteByte(',')
			}
			encodedName, err := json.Marshal(name)
			if err != nil {
				return nil, err
			}
			encodedValue, err := marshalColumnTypeDetails(d.Properties[name], nil, stack)
			if err != nil {
				return nil, err
			}
			buffer.Write(encodedName)
			buffer.WriteByte(':')
			buffer.Write(encodedValue)
		}
		buffer.WriteByte('}')
	}
	if d.Element != nil {
		encodedElement, err := marshalColumnTypeDetails(*d.Element, d.Element, stack)
		if err != nil {
			return nil, err
		}
		buffer.WriteString(`,"element":`)
		buffer.Write(encodedElement)
	}
	buffer.WriteByte('}')
	return buffer.Bytes(), nil
}

func decodeStrictJSON(data []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("unexpected trailing JSON value")
		}
		return err
	}
	return nil
}

func orderedTableMetadata(data []byte) ([]string, map[string][]string) {
	return orderedTableMetadataForField(data, "tables")
}

func orderedTableMetadataForField(data []byte, fieldName string) ([]string, map[string][]string) {
	fields, err := decodeOrderedJSONObject(data)
	if err != nil {
		return nil, nil
	}
	for _, field := range fields {
		if field.Name != fieldName {
			continue
		}
		tables, err := decodeOrderedJSONObject(field.Value)
		if err != nil {
			return nil, nil
		}
		order := make([]string, 0, len(tables))
		columns := make(map[string][]string, len(tables))
		for _, table := range tables {
			order = append(order, table.Name)
			columnFields, err := decodeOrderedJSONObject(table.Value)
			if err != nil {
				continue
			}
			columnOrder := make([]string, 0, len(columnFields))
			for _, column := range columnFields {
				columnOrder = append(columnOrder, column.Name)
			}
			columns[table.Name] = columnOrder
		}
		return order, columns
	}
	return nil, nil
}

func decodeOrderedJSONObject(data []byte) ([]orderedJSONField, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != '{' {
		return nil, fmt.Errorf("expected JSON object")
	}
	var fields []orderedJSONField
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		name, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("expected JSON object field")
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		fields = append(fields, orderedJSONField{Name: name, Value: value})
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	return fields, nil
}
