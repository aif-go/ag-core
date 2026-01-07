package erm

import "encoding/xml"

type Diagram struct {
	XMLName xml.Name `xml:"diagram"`
	Dictionary Dictionary `xml:"dictionary"`
	Contents Contents `xml:"contents"`
	DbType string
}

type Dictionary struct{
	XMLName xml.Name `xml:"dictionary"`
	Words []*Word `xml:"word"`
}

type Word struct{
	XMLName xml.Name `xml:"word"`
	ID string `xml:"id"`
	Length string `xml:"length"`
	Decimal string `xml:"decimal"`
	Array string `xml:"array"`
	ArrayDimension string `xml:"array_dimension"`
	Unsigned bool `xml:"unsigned"`
	Description string `xml:"description"`
	LogicalName string `xml:"logical_name"`
	PhysicalName string `xml:"physical_name"`
	Type string `xml:"type"`
}

type Contents struct{
	XMLName xml.Name `xml:"contents"`
	Tables []*Table `xml:"table"`

	// table

}

type Table struct{
	XMLName xml.Name `xml:"table"`
	Columns Columns `xml:"columns"`
	Indexs Indexs `xml:"indexes"`
	UniqueIndexes UniqueIndexes `xml:"complex_unique_key_list"`
	TableName string `xml:"physical_name"`
	TableProperties TableProperties `xml:"table_properties"`
}

type TableProperties struct{
	EnCode string `xml:"character_set"`
	Sort string `xml:"collation"`
	Engine string `xml:"storage_engine"`
}

type Columns struct{
	XMLName xml.Name `xml:"columns"`
	NormalColumns []*NormalColumn `xml:"normal_column"`
}

type Indexs struct{
	XMLName xml.Name `xml:"indexes"`
	Indexs []*Index `xml:"inidex"`
}

// UniqueIndexes 唯一索引集合
type UniqueIndexes struct{
	XMLName xml.Name `xml:"complex_unique_key_list"`
	UniqueIndexes [] *UniqueIndex `xml:"complex_unique_key"`
}

// UniqueIndex 唯一索引明细
type UniqueIndex struct{
	XMLName xml.Name `xml:"complex_unique_key"`
	ID string `xml:"id"`
	Name string `xml:"name"`
	IndexColumns IndexColumns `xml:"columns"`
}

type NormalColumn struct{
	XMLName xml.Name `xml:"normal_column"`
	Sequence Sequence `xml:"sequence"`
	WordId string `xml:"word_id"`
	ID string `xml:"id"`
	ReferencedColumn string `xml:"referenced_column"`
	Description string `xml:"description"`
	UniquekeyName string `xml:"unique_key_name"`
	LogicalName string `xml:"logical_name"`
	PhysicalName string `xml:"physical_name"`
	Type string `xml:"type"`
	Constraint string `xml:"constraint"`
	DefaultValue string `xml:"default_value"`
	AutoIncrement bool `xml:"auto_increment"`
	ForeignKey bool `xml:"foreign_key"`
	NotNull bool `xml:"not_null"`
	PrimaryKey bool `xml:"primary_key"`
	Uniquekey bool `xml:"unique_key"`
	// CharacterSet string `xml:"character_set"`
	// Collation string `xml:"collation"`
}

type Index struct{
	XMLName xml.Name `xml:"inidex"`
	// `xml:full_text`
	NonUnique bool `xml:"non_unique"`
	Name string `xml:"name"`
	Type string `xml:"type"`
    IndexColumns IndexColumns `xml:"columns"`
}

type IndexColumns struct{
	XMLName xml.Name `xml:"columns"`
	IndexColumn []*IndexColumn `xml:"column"`
}

type IndexColumn struct{
	XMLName xml.Name `xml:"column"`
	ID string `xml:"id"`
	Desc string `xml:"desc"`
}

type Sequence struct{

}