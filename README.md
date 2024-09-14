(WIP)

# Easy Storage
A new way for your file storaging. Away from cloud!



### Testing

Posting a file
```bash
curl -X POST \
  http://localhost:8080/upload \
  -F 'file=@path-to-file' \
  -F 'key=file-key' -v
```

Retrieving a file
```bash
curl -X GET \
  http://localhost:8080/download/{file-key}
```

Retrieving signed url for a file
```bash
curl -X GET \
  http://localhost:8080/signed/{file-key}
```
