Transform a mysqldump into a json file, the JSON file into a dot file, and the dot file into a ERD.

```bash
go install github.com/tjgurwara99/sqljson/cmd/sqljsondump@v0.0.3
go install github.com/tjgurwara99/sqljson/cmd/sqljsondot@v0.0.3

mysqldump --no-data database | sqljsondump | sqljsondot | dot -Tpng -o database.png && open database.png
```
