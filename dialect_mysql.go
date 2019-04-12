/*
 * Copyright 2018. bigpigeon. All rights reserved.
 * Use of this source code is governed by a MIT style
 * license that can be found in the LICENSE file.
 */

package toyorm

import (
	"database/sql"
	"fmt"
	"strings"
)

type MySqlDialect struct {
	DefaultDialect
}

var mysqlKeywordMap map[string]struct{}

func init() {
	mysqlKeywordMap = map[string]struct{}{}
	keywords := []string{"ACCESSIBLE", "ACCOUNT", "ACTION", "ADD", "ADMIN", "AFTER", "AGAINST", "AGGREGATE", "ALGORITHM", "ALL", "ALTER", "ALWAYS", "ANALYSE", "ANALYZE", "AND", "ANY", "AS", "ASC", "ASCII", "ASENSITIVE", "AT", "AUTHORS", "AUTOEXTEND_SIZE", "AUTO_INCREMENT", "AVG", "AVG_ROW_LENGTH", "BACKUP", "BEFORE", "BEGIN", "BETWEEN", "BIGINT", "BINARY", "BINLOG", "BIT", "BLOB", "BLOCK", "BOOL", "BOOLEAN", "BOTH", "BTREE", "BUCKETS", "BY", "BYTE", "CACHE", "CALL", "CASCADE", "CASCADED", "CASE", "CATALOG_NAME", "CHAIN", "CHANGE", "CHANGED", "CHANNEL", "CHAR", "CHARACTER", "CHARSET", "CHECK", "CHECKSUM", "CIPHER", "CLASS_ORIGIN", "CLIENT", "CLONE", "CLOSE", "COALESCE", "CODE", "COLLATE", "COLLATION", "COLUMN", "COLUMNS", "COLUMN_FORMAT", "COLUMN_NAME", "COMMENT", "COMMIT", "COMMITTED", "COMPACT", "COMPLETION", "COMPONENT", "COMPRESSED", "COMPRESSION", "CONCURRENT", "CONDITION", "CONNECTION", "CONSISTENT", "CONSTRAINT", "CONSTRAINT_CATALOG", "CONSTRAINT_NAME", "CONSTRAINT_SCHEMA", "CONTAINS", "CONTEXT", "CONTINUE", "CONTRIBUTORS", "CONVERT", "CPU", "CREATE", "CROSS", "CUBE", "CUME_DIST", "CURRENT", "CURRENT_DATE", "CURRENT_TIME", "CURRENT_TIMESTAMP", "CURRENT_USER", "CURSOR", "CURSOR_NAME", "DATA", "DATABASE", "DATABASES", "DATAFILE", "DATE", "DATETIME", "DAY", "DAY_HOUR", "DAY_MICROSECOND", "DAY_MINUTE", "DAY_SECOND", "DEALLOCATE", "DEC", "DECIMAL", "DECLARE", "DEFAULT", "DEFAULT_AUTH", "DEFINER", "DEFINITION", "DELAYED", "DELAY_KEY_WRITE", "DELETE", "DENSE_RANK", "DESC", "DESCRIBE", "DESCRIPTION", "DES_KEY_FILE", "DETERMINISTIC", "DIAGNOSTICS", "DIRECTORY", "DISABLE", "DISCARD", "DISK", "DISTINCT", "DISTINCTROW", "DIV", "DO", "DOUBLE", "DROP", "DUAL", "DUMPFILE", "DUPLICATE", "DYNAMIC", "EACH", "ELSE", "ELSEIF", "EMPTY", "ENABLE", "ENCLOSED", "ENCRYPTION", "END", "ENDS", "ENGINE", "ENGINES", "ENUM", "ERROR", "ERRORS", "ESCAPE", "ESCAPED", "EVENT", "EVENTS", "EVERY", "EXCEPT", "EXCHANGE", "EXCLUDE", "EXECUTE", "EXISTS", "EXIT", "EXPANSION", "EXPIRE", "EXPLAIN", "EXPORT", "EXTENDED", "EXTENT_SIZE", "FALSE", "FAST", "FAULTS", "FETCH", "FIELDS", "FILE", "FILE_BLOCK_SIZE", "FILTER", "FIRST", "FIRST_VALUE", "FIXED", "FLOAT", "FLOAT4", "FLOAT8", "FLUSH", "FOLLOWING", "FOLLOWS", "FOR", "FORCE", "FOREIGN", "FORMAT", "FOUND", "FRAC_SECOND", "FROM", "FULL", "FULLTEXT", "FUNCTION", "GENERAL", "GENERATED", "GEOMCOLLECTION", "GEOMETRY", "GEOMETRYCOLLECTION", "GET", "GET_FORMAT", "GET_MASTER_PUBLIC_KEY", "GLOBAL", "GRANT", "GRANTS", "GROUP", "GROUPING", "GROUPS", "GROUP_REPLICATION", "HANDLER", "HASH", "HAVING", "HELP", "HIGH_PRIORITY", "HISTOGRAM", "HISTORY", "HOST", "HOSTS", "HOUR", "HOUR_MICROSECOND", "HOUR_MINUTE", "HOUR_SECOND", "IDENTIFIED", "IF", "IGNORE", "IGNORE_SERVER_IDS", "IMPORT", "IN", "INDEX", "INDEXES", "INFILE", "INITIAL_SIZE", "INNER", "INNOBASE", "INNODB", "INOUT", "INSENSITIVE", "INSERT", "INSERT_METHOD", "INSTALL", "INSTANCE", "INT", "INT1", "INT2", "INT3", "INT4", "INT8", "INTEGER", "INTERVAL", "INTO", "INVISIBLE", "INVOKER", "IO", "IO_AFTER_GTIDS", "IO_BEFORE_GTIDS", "IO_THREAD", "IPC", "IS", "ISOLATION", "ISSUER", "ITERATE", "JOIN", "JSON", "JSON_TABLE", "KEY", "KEYS", "KEY_BLOCK_SIZE", "KILL", "LAG", "LANGUAGE", "LAST", "LAST_VALUE", "LEAD", "LEADING", "LEAVE", "LEAVES", "LEFT", "LESS", "LEVEL", "LIKE", "LIMIT", "LINEAR", "LINES", "LINESTRING", "LIST", "LOAD", "LOCAL", "LOCALTIME", "LOCALTIMESTAMP", "LOCK", "LOCKED", "LOCKS", "LOGFILE", "LOGS", "LONG", "LONGBLOB", "LONGTEXT", "LOOP", "LOW_PRIORITY", "MASTER", "MASTER_AUTO_POSITION", "MASTER_BIND", "MASTER_CONNECT_RETRY", "MASTER_DELAY", "MASTER_HEARTBEAT_PERIOD", "MASTER_HOST", "MASTER_LOG_FILE", "MASTER_LOG_POS", "MASTER_PASSWORD", "MASTER_PORT", "MASTER_PUBLIC_KEY_PATH", "MASTER_RETRY_COUNT", "MASTER_SERVER_ID", "MASTER_SSL", "MASTER_SSL_CA", "MASTER_SSL_CAPATH", "MASTER_SSL_CERT", "MASTER_SSL_CIPHER", "MASTER_SSL_CRL", "MASTER_SSL_CRLPATH", "MASTER_SSL_KEY", "MASTER_SSL_VERIFY_SERVER_CERT", "MASTER_TLS_VERSION", "MASTER_USER", "MATCH", "MAXVALUE", "MAX_CONNECTIONS_PER_HOUR", "MAX_QUERIES_PER_HOUR", "MAX_ROWS", "MAX_SIZE", "MAX_UPDATES_PER_HOUR", "MAX_USER_CONNECTIONS", "MEDIUM", "MEDIUMBLOB", "MEDIUMINT", "MEDIUMTEXT", "MEMORY", "MERGE", "MESSAGE_TEXT", "MICROSECOND", "MIDDLEINT", "MIGRATE", "MINUTE", "MINUTE_MICROSECOND", "MINUTE_SECOND", "MIN_ROWS", "MOD", "MODE", "MODIFIES", "MODIFY", "MONTH", "MULTILINESTRING", "MULTIPOINT", "MULTIPOLYGON", "MUTEX", "MYSQL_ERRNO", "NAME", "NAMES", "NATIONAL", "NATURAL", "NCHAR", "NDB", "NDBCLUSTER", "NESTED", "NEVER", "NEW", "NEXT", "NO", "NODEGROUP", "NONE", "NOT", "NOWAIT", "NO_WAIT", "NO_WRITE_TO_BINLOG", "NTH_VALUE", "NTILE", "NULL", "NULLS", "NUMBER", "NUMERIC", "NVARCHAR", "OF", "OFFSET", "OLD_PASSWORD", "ON", "ONE", "ONE_SHOT", "ONLY", "OPEN", "OPTIMIZE", "OPTIMIZER_COSTS", "OPTION", "OPTIONALLY", "OPTIONS", "OR", "ORDER", "ORDINALITY", "OTHERS", "OUT", "OUTER", "OUTFILE", "OVER", "OWNER", "PACK_KEYS", "PAGE", "PARSER", "PARSE_GCOL_EXPR", "PARTIAL", "PARTITION", "PARTITIONING", "PARTITIONS", "PASSWORD", "PATH", "PERCENT_RANK", "PERSIST", "PERSIST_ONLY", "PHASE", "PLUGIN", "PLUGINS", "PLUGIN_DIR", "POINT", "POLYGON", "PORT", "PRECEDES", "PRECEDING", "PRECISION", "PREPARE", "PRESERVE", "PREV", "PRIMARY", "PRIVILEGES", "PROCEDURE", "PROCESS", "PROCESSLIST", "PROFILE", "PROFILES", "PROXY", "PURGE", "QUARTER", "QUERY", "QUICK", "RANGE", "RANK", "READ", "READS", "READ_ONLY", "READ_WRITE", "REAL", "REBUILD", "RECOVER", "RECURSIVE", "REDOFILE", "REDO_BUFFER_SIZE", "REDUNDANT", "REFERENCE", "REFERENCES", "REGEXP", "RELAY", "RELAYLOG", "RELAY_LOG_FILE", "RELAY_LOG_POS", "RELAY_THREAD", "RELEASE", "RELOAD", "REMOTE", "REMOVE", "RENAME", "REORGANIZE", "REPAIR", "REPEAT", "REPEATABLE", "REPLACE", "REPLICATE_DO_DB", "REPLICATE_DO_TABLE", "REPLICATE_IGNORE_DB", "REPLICATE_IGNORE_TABLE", "REPLICATE_REWRITE_DB", "REPLICATE_WILD_DO_TABLE", "REPLICATE_WILD_IGNORE_TABLE", "REPLICATION", "REQUIRE", "RESET", "RESIGNAL", "RESOURCE", "RESPECT", "RESTART", "RESTORE", "RESTRICT", "RESUME", "RETURN", "RETURNED_SQLSTATE", "RETURNS", "REUSE", "REVERSE", "REVOKE", "RIGHT", "RLIKE", "ROLE", "ROLLBACK", "ROLLUP", "ROTATE", "ROUTINE", "ROW", "ROWS", "ROW_COUNT", "ROW_FORMAT", "ROW_NUMBER", "RTREE", "SAVEPOINT", "SCHEDULE", "SCHEMA", "SCHEMAS", "SCHEMA_NAME", "SECOND", "SECOND_MICROSECOND", "SECURITY", "SELECT", "SENSITIVE", "SEPARATOR", "SERIAL", "SERIALIZABLE", "SERVER", "SESSION", "SET", "SHARE", "SHOW", "SHUTDOWN", "SIGNAL", "SIGNED", "SIMPLE", "SKIP", "SLAVE", "SLOW", "SMALLINT", "SNAPSHOT", "SOCKET", "SOME", "SONAME", "SOUNDS", "SOURCE", "SPATIAL", "SPECIFIC", "SQL", "SQLEXCEPTION", "SQLSTATE", "SQLWARNING", "SQL_AFTER_GTIDS", "SQL_AFTER_MTS_GAPS", "SQL_BEFORE_GTIDS", "SQL_BIG_RESULT", "SQL_BUFFER_RESULT", "SQL_CACHE", "SQL_CALC_FOUND_ROWS", "SQL_NO_CACHE", "SQL_SMALL_RESULT", "SQL_THREAD", "SQL_TSI_DAY", "SQL_TSI_FRAC_SECOND", "SQL_TSI_HOUR", "SQL_TSI_MINUTE", "SQL_TSI_MONTH", "SQL_TSI_QUARTER", "SQL_TSI_SECOND", "SQL_TSI_WEEK", "SQL_TSI_YEAR", "SRID", "SSL", "STACKED", "START", "STARTING", "STARTS", "STATS_AUTO_RECALC", "STATS_PERSISTENT", "STATS_SAMPLE_PAGES", "STATUS", "STOP", "STORAGE", "STORED", "STRAIGHT_JOIN", "STRING", "SUBCLASS_ORIGIN", "SUBJECT", "SUBPARTITION", "SUBPARTITIONS", "SUPER", "SUSPEND", "SWAPS", "SWITCHES", "SYSTEM", "TABLE", "TABLES", "TABLESPACE", "TABLE_CHECKSUM", "TABLE_NAME", "TEMPORARY", "TEMPTABLE", "TERMINATED", "TEXT", "THAN", "THEN", "THREAD_PRIORITY", "TIES", "TIME", "TIMESTAMP", "TIMESTAMPADD", "TIMESTAMPDIFF", "TINYBLOB", "TINYINT", "TINYTEXT", "TO", "TRAILING", "TRANSACTION", "TRIGGER", "TRIGGERS", "TRUE", "TRUNCATE", "TYPE", "TYPES", "UNBOUNDED", "UNCOMMITTED", "UNDEFINED", "UNDO", "UNDOFILE", "UNDO_BUFFER_SIZE", "UNICODE", "UNINSTALL", "UNION", "UNIQUE", "UNKNOWN", "UNLOCK", "UNSIGNED", "UNTIL", "UPDATE", "UPGRADE", "USAGE", "USE", "USER", "USER_RESOURCES", "USE_FRM", "USING", "UTC_DATE", "UTC_TIME", "VALIDATION", "VALUES", "VARBINARY", "VARCHAR", "VARCHARACTER", "VARIABLES", "VARYING", "VCPU", "VIEW", "VIRTUAL", "VISIBLE", "WAIT", "WARNINGS", "WEEK", "WEIGHT_STRING", "WHEN", "WHERE", "WHILE", "WINDOW", "WITH", "WITHOUT", "WORK", "WRAPPER", "WRITE", "X509", "XA", "XID", "XML", "XOR", "YEAR", "YEAR_MONTH", "ZEROFILL"}
	for _, word := range keywords {
		mysqlKeywordMap[word] = struct{}{}
	}
}

func (dia MySqlDialect) SaveExecutor(db Executor, exec ExecValue, debugPrinter func(ExecValue, error)) (sql.Result, error) {
	query := exec.Query()
	result, err := db.Exec(query, exec.Args()...)
	debugPrinter(exec, err)
	return result, err
}

func (dia MySqlDialect) HasTable(model *Model) ExecValue {
	return DefaultExec{
		"SELECT count(*) FROM INFORMATION_SCHEMA.TABLES WHERE  table_schema = (SELECT DATABASE()) AND table_name = ?",
		[]interface{}{model.Name},
	}
}

func (dia MySqlDialect) CreateTable(model *Model, foreign map[string]ForeignKey) (execlist []ExecValue) {
	// lazy init model
	strList := []string{}
	// use to create foreign definition

	for _, sqlField := range model.GetSqlFields() {

		s := fmt.Sprintf("%s %s", sqlField.Column(), sqlField.SqlType())
		if sqlField.AutoIncrement() {
			s += " AUTO_INCREMENT"
		}
		if _default := sqlField.Default(); _default != "" {
			s += " DEFAULT " + _default
		}
		for k, v := range sqlField.Attrs() {
			if v == "" {
				s += " " + k
			} else {
				s += " " + fmt.Sprintf("%s=%s", k, v)
			}
		}
		strList = append(strList, s)
	}
	var primaryStrList []string
	for _, p := range model.GetPrimary() {
		primaryStrList = append(primaryStrList, p.Column())
	}
	strList = append(strList, fmt.Sprintf("PRIMARY KEY(%s)", strings.Join(primaryStrList, ",")))

	for name, key := range foreign {
		f := model.GetFieldWithName(name)
		strList = append(strList,
			fmt.Sprintf("FOREIGN KEY (%s) REFERENCES `%s`(%s)", f.Column(), key.Model.Name, key.Field.Column()),
		)
	}

	sqlStr := fmt.Sprintf("CREATE TABLE `%s` (%s)",
		model.Name,
		strings.Join(strList, ","),
	)
	execlist = append(execlist, DefaultExec{sqlStr, nil})

	indexStrList := []string{}
	for key, fieldList := range model.GetIndexMap() {
		fieldStrList := []string{}
		for _, f := range fieldList {
			fieldStrList = append(fieldStrList, f.Column())
		}
		indexStrList = append(indexStrList, fmt.Sprintf("CREATE INDEX %s ON `%s`(%s)", key, model.Name, strings.Join(fieldStrList, ",")))
	}
	uniqueIndexStrList := []string{}
	for key, fieldList := range model.GetUniqueIndexMap() {
		fieldStrList := []string{}
		for _, f := range fieldList {
			fieldStrList = append(fieldStrList, f.Column())
		}
		uniqueIndexStrList = append(uniqueIndexStrList, fmt.Sprintf("CREATE UNIQUE INDEX %s ON `%s`(%s)", key, model.Name, strings.Join(fieldStrList, ",")))
	}
	for _, indexStr := range indexStrList {
		execlist = append(execlist, DefaultExec{indexStr, nil})
	}
	for _, indexStr := range uniqueIndexStrList {
		execlist = append(execlist, DefaultExec{indexStr, nil})
	}
	return
}

// replace will failure when have foreign key
func (dia MySqlDialect) SaveExec(model *Model, columnValues []ColumnNameValue) ExecValue {
	var exec ExecValue = DefaultExec{}
	fieldStr, qStr, args := insertValuesFormat(model, columnValues)
	exec = exec.Append(
		fmt.Sprintf("INSERT INTO `%s`(%s) VALUES(%s)", model.Name, fieldStr, qStr),
		args...,
	)

	var recordList []string
	for _, r := range columnValues {
		switch r.Name() {
		case "CreatedAt":
			//recordList = append(recordList, fmt.Sprintf("%[1]s = VALUES(%[1]s)", r.Column()))
		case "Cas":
			recordList = append(recordList, fmt.Sprintf("%[1]s = IF(%[1]s = VALUES(%[1]s) - 1, VALUES(%[1]s) , \"update failure\")", dia.keywordProcess(r.Column(), model.Name)))
		default:
			recordList = append(recordList, fmt.Sprintf("%[1]s = VALUES(%[1]s)", dia.keywordProcess(r.Column(), model.Name)))
		}

	}
	exec = exec.Append(fmt.Sprintf(" ON DUPLICATE KEY UPDATE %s",
		strings.Join(recordList, ","),
	))

	return exec
}

func (dia MySqlDialect) keywordProcess(s string, modelName string) string {
	_, ok := mysqlKeywordMap[strings.ToUpper(s)]
	if ok {
		return modelName + "." + s
	}
	return s
}
