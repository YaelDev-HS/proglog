# Prolog


## Ejecutar el log.proto (o cualquier archivo de protobuf)
```
protoc --proto_path=. --go_out=. --go_opt=paths=source_relative api/v1/*.proto
```

En caso de actualizar algun campo, reecompilas el proyecto .proto.

# IMPORTANTE
los IDs son inmutables, por lo que no se pueden cambiar.