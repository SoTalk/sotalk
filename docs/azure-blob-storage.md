# Azure Blob Storage Integration

The media upload system now supports Azure Blob Storage for production-grade file storage.

## Features

- **Dual Storage Support**: Choose between Azure Blob Storage (production) or local filesystem (development)
- **Seamless Integration**: Same API for both storage providers
- **Production Ready**: Scalable, secure, and reliable cloud storage
- **Easy Configuration**: Simple environment variable configuration

## Configuration

### Using Azure Blob Storage

1. Create an Azure Storage Account:
   ```bash
   # Using Azure CLI
   az storage account create \
     --name youraccountname \
     --resource-group your-resource-group \
     --location eastus \
     --sku Standard_LRS
   ```

2. Get your storage account key:
   ```bash
   az storage account keys list \
     --account-name youraccountname \
     --resource-group your-resource-group
   ```

3. Create a container:
   ```bash
   az storage container create \
     --name media \
     --account-name youraccountname \
     --account-key your-account-key
   ```

4. Update your `.env` file:
   ```env
   STORAGE_PROVIDER=azure
   AZURE_STORAGE_ACCOUNT_NAME=youraccountname
   AZURE_STORAGE_ACCOUNT_KEY=your-account-key
   AZURE_STORAGE_CONTAINER=media
   ```

### Using Local Storage (Development)

For local development, use local filesystem storage:

```env
STORAGE_PROVIDER=local
STORAGE_LOCAL_PATH=./uploads
```

## How It Works

### Upload Flow

1. Client uploads file via multipart form
2. Media service validates file (size, type, quota)
3. Storage service uploads to Azure Blob or local filesystem
4. Media metadata saved to database with blob URL
5. Client receives blob URL for access

### File Structure

**Azure Blob Storage:**
```
container: media
├── {user-id}/
│   ├── {uuid}-filename.jpg
│   ├── {uuid}-document.pdf
│   └── {uuid}-voice.ogg
```

**Local Storage:**
```
./uploads/
├── {user-id}/
│   ├── {uuid}-filename.jpg
│   ├── {uuid}-document.pdf
│   └── {uuid}-voice.ogg
```

## API Endpoints

### Upload Media
```http
POST /api/v1/media/upload
Content-Type: multipart/form-data
Authorization: Bearer {token}

file: [binary data]
```

### Upload Voice Message
```http
POST /api/v1/media/voice
Content-Type: multipart/form-data
Authorization: Bearer {token}

file: [audio file]
```

### Upload File/Document
```http
POST /api/v1/media/file
Content-Type: multipart/form-data
Authorization: Bearer {token}

file: [document]
```

### Get Storage Info
```http
GET /api/v1/media/storage
Authorization: Bearer {token}
```

Response:
```json
{
  "total_bytes": 52428800,
  "total_mb": 50,
  "total_gb": 0.048828125,
  "max_gb": 1,
  "percent_used": 4.88,
  "available_gb": 0.951171875
}
```

## Storage Limits

- **Max File Size**: 50 MB per file
- **Max Storage Per User**: 1 GB total
- **Supported Formats**: Images, videos, audio, documents (PDF, DOC, DOCX)

## Security

### Azure Blob Storage
- Files stored with unique UUID prefixes
- Account key stored securely in environment variables
- Container access can be set to private
- Optional: Use Azure CDN for content delivery

### Local Storage
- Files stored in designated directory
- Directory permissions managed by OS
- Not recommended for production

## Performance

- **Azure Blob**: Globally distributed, CDN-ready
- **Local**: Fast for development, limited by disk I/O

## Cost Considerations

**Azure Blob Storage Pricing (example):**
- Storage: ~$0.018 per GB per month (hot tier)
- Bandwidth: Free ingress, ~$0.087 per GB egress
- Operations: ~$0.004 per 10,000 write operations

**Estimated monthly cost for 1000 active users:**
- Storage (500 MB avg per user): ~$9/month
- Bandwidth (10 GB downloads): ~$0.87/month
- **Total**: ~$10/month

## Troubleshooting

### Connection Issues
```bash
# Test Azure Blob connection
az storage blob list \
  --account-name youraccountname \
  --account-key your-account-key \
  --container-name media
```

### Permission Issues
Ensure your storage account key has the correct permissions:
- Read access
- Write access
- Delete access

### Logs
Check application logs for storage errors:
```
INFO: Azure Blob Storage initialized
INFO: File uploaded to Azure Blob, blob_name=...
ERROR: Failed to upload to Azure Blob: ...
```

## Migration

### Moving from Local to Azure

1. Update environment variables to use Azure
2. Restart application
3. Optional: Migrate existing files using `azcopy`:
   ```bash
   azcopy copy './uploads/*' \
     'https://youraccountname.blob.core.windows.net/media' \
     --recursive
   ```

## Best Practices

1. **Production**: Always use Azure Blob Storage
2. **Development**: Use local storage for faster iteration
3. **Backups**: Enable Azure Blob versioning and soft delete
4. **CDN**: Use Azure CDN for global content delivery
5. **Monitoring**: Enable Azure Storage Analytics
6. **Security**: Rotate storage account keys regularly
7. **Access**: Use SAS tokens for temporary access instead of full keys

## Next Steps

- [ ] Implement CDN integration
- [ ] Add image compression/resizing
- [ ] Add virus scanning for uploads
- [ ] Implement presigned URLs for private access
- [ ] Add S3-compatible storage support
