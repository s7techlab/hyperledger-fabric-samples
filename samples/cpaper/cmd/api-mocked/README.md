# Commercial paper API - mocked version

```
# go run main.go
```

## Example usage 

### Issue commercial paper

```
API_HOST=http://localhost:8081 ./issue.sh payload_cpaper_1_issue.json
```

```
API_HOST=http://localhost:8081 ./issue.sh payload_cpaper_2_issue.json
```

### List issued papers 

```
API_HOST=http://localhost:8081 ./list.sh
```