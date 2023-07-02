package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"

	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/tjgurwara99/sqljson/internal/relationships"
	"github.com/tjgurwara99/sqljson/internal/types"
	"golang.org/x/exp/maps"

	_ "github.com/pingcap/parser/test_driver"

	"os"

	pg_query "github.com/pganalyze/pg_query_go/v4"
)

// Transform returns a function that will read a MySQL dump from r and write a JSON description to w.
func Transform(r io.Reader, w io.Writer) error {
	p := parser.New()

	dump, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	statements, _, err := p.Parse(string(dump), "", "")
	if err != nil {
		return err
	}

	tables := make(map[string]*types.CreateTable)

	for _, statement := range statements {
		if create, ok := statement.(*ast.CreateTableStmt); ok {
			tableName := create.Table.Name.String()

			createTable := &types.CreateTable{
				Columns: make(map[string]*types.CreateColumn),
			}

			for _, col := range create.Cols {
				columnName := col.Name.String()

				createColumn := &types.CreateColumn{
					Type: col.Tp.InfoSchemaStr(),
				}
				createTable.Columns[columnName] = createColumn
			}

			tables[tableName] = createTable
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	return enc.Encode(&tables)
}

func TransformPostgres(r io.Reader, w io.Writer) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	statements, err := pg_query.Parse(string(data))
	if err != nil {
		return err
	}
	tables := make(map[string]*relationships.CreateTable)

	for _, statement := range statements.Stmts {
		create := statement.Stmt.GetCreateStmt()
		relation := make(map[string]string)
		if create == nil {
			// must be a relation statement
			alt := statement.Stmt.GetAlterTableStmt()
			if alt == nil {
				continue
			}
			tableName := alt.Relation.Relname
			left := alt.Cmds[0].GetAlterTableCmd()
			con := left.Def.GetConstraint()
			relation[con.FkAttrs[0].GetString_().Sval] = con.Pktable.Relname
			tbl, ok := tables[tableName]
			if !ok {
				tables[tableName] = &relationships.CreateTable{
					Relationships: relation,
				}
				continue
			}
			maps.Copy(tbl.Relationships, relation)
			continue
		}
		tableName := create.Relation.Relname

		createTable := &types.CreateTable{
			Columns: make(map[string]*types.CreateColumn),
		}

		for _, col := range create.TableElts {
			c := col.GetColumnDef()
			if len(c.Constraints) != 0 {
				_ = c.Constraints[0].GetConstraint()
			}
			n := len(c.GetTypeName().GetNames())
			ss, ok := c.GetTypeName().GetNames()[n-1].Node.(*pg_query.Node_String_)
			if !ok {
				log.Printf("something went wrong here: %s", err)
				continue
			}
			createColumn := &types.CreateColumn{
				Type: ss.String_.Sval,
			}
			createTable.Columns[c.Colname] = createColumn
		}

		tables[tableName] = &relationships.CreateTable{
			CreateTable:   *createTable,
			Relationships: relation,
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", " ")
	return enc.Encode(&tables)
}

func main() {
	sqlType := flag.String("type", "mysql", "sql variant to parse")
	flag.Parse()
	switch *sqlType {
	case "mysql":
		if err := Transform(os.Stdin, os.Stdout); err != nil {
			log.Fatal(err)
		}
	case "postgres":
		if err := TransformPostgres(os.Stdin, os.Stdout); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown sql variant provided: %s", *sqlType)
	}
}
