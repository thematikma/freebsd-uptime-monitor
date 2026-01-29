# Proxy Deployment Troubleshooting

## White/Blank Page Behind Traefik Proxy

If you're seeing a white page with GET requests in logs, this indicates the HTML is loading but assets (CSS/JS) are not. Here are the fixes:

### 1. Rebuild Frontend with Proxy Configuration

The updated SvelteKit configuration now supports proxy deployment:

```bash
cd frontend
npm install
npm run build
```

### 2. Update Backend Configuration

The Go backend now includes:
- Proper MIME type headers
- CORS support for development
- SPA routing fallbacks
- Static asset serving with correct paths

### 3. Traefik Configuration

Ensure your Traefik setup includes proper headers:

```yaml
http:
  middlewares:
    uptime-headers:
      headers:
        customRequestHeaders:
          X-Forwarded-Proto: "https"
          X-Forwarded-For: ""
```

### 4. Common Issues & Solutions

**Problem**: Assets not loading (404 errors in browser console)
**Solution**: Check that `/assets` and `/_app` paths are correctly routed to the backend

**Problem**: API calls failing
**Solution**: Ensure `/api` paths are routed to the backend, not served as static files

**Problem**: Page refreshing returns 404
**Solution**: SPA fallback is now configured - all non-API routes serve index.html

### 5. Testing

1. **Check browser console** for 404 errors on assets
2. **Verify API endpoints** work: `curl https://yoursite.com/api/v1/dashboard`
3. **Test WebSocket** connection for real-time updates
4. **Check headers** in browser dev tools Network tab

### 6. Production Checklist

- [ ] Frontend built with `npm run build`
- [ ] Backend compiled with updated proxy support
- [ ] Traefik routes `/api` to backend
- [ ] Traefik serves static files from backend
- [ ] WebSocket `/ws` endpoint configured
- [ ] HTTPS certificates working
- [ ] Headers middleware configured

### 7. Debug Commands

```bash
# Check if assets are accessible
curl -I https://yoursite.com/assets/

# Check API endpoint
curl https://yoursite.com/api/v1/dashboard

# Check WebSocket (using websocat if available)
websocat wss://yoursite.com/ws
```