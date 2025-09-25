# 🚀 Production Deployment Guide

This comprehensive guide ensures your application is production-ready and prevents API-related issues.

## 📋 Pre-Deployment Checklist

### ✅ Environment Configuration
- [ ] Copy `.env.production.template` to `.env.production` with actual values
- [ ] Set `NODE_ENV=production`
- [ ] Configure `NEXT_PUBLIC_API_URL` to production API server
- [ ] Set appropriate `API_TIMEOUT` and `API_RETRIES` values
- [ ] Enable monitoring with `MONITORING_ENABLED=true`
- [ ] Configure logging level (`LOGGING_LEVEL=warn` recommended)

### ✅ Security Configuration
- [ ] Enable HTTPS enforcement (`ENFORCE_HTTPS=true`)
- [ ] Configure allowed origins (`ALLOWED_ORIGINS`)
- [ ] Ensure API URL uses HTTPS
- [ ] Review security headers configuration

### ✅ Performance Optimization
- [ ] Enable caching (`CACHE_ENABLED=true`)
- [ ] Configure cache TTL and max size
- [ ] Enable compression (`ENABLE_COMPRESSION=true`)
- [ ] Enable static asset caching

### ✅ Feature Flags
- [ ] Disable debug mode (`ENABLE_DEBUG_MODE=false`)
- [ ] Disable experimental features (`ENABLE_EXPERIMENTAL_FEATURES=false`)
- [ ] Disable API mocking (`ENABLE_API_MOCKING=false`)
- [ ] Enable endpoint validation (`VALIDATE_ENDPOINTS=true`)

### ✅ Monitoring Setup (Optional but Recommended)
- [ ] Configure Sentry for error monitoring (`SENTRY_DSN`)
- [ ] Set up New Relic APM (`NEW_RELIC_LICENSE_KEY`)
- [ ] Configure monitoring service webhooks
- [ ] Test error reporting in staging

## 🔧 Validation Commands

Before deployment, run these validation commands:

```bash
# Install dependencies
npm ci

# Run environment validation
npm run validate:env

# Run API endpoint validation
npm run validate:api

# Run full production readiness check
npm run validate:production

# Build production bundle
npm run build

# Test production build locally
npm run start
```

## 📊 Production Readiness Validation

### Automated Validation Scripts

Add these scripts to your `package.json`:

```json
{
  "scripts": {
    "validate:env": "node -e \"require('./src/config/production').validateEnvironmentVariables()\"",
    "validate:api": "node -e \"require('./src/utils/apiValidation').validateCriticalEndpoints().then(r => console.log(r))\"",
    "validate:production": "node -e \"require('./src/config/production').checkProductionReadiness().then(r => console.log(r))\"",
    "health-check": "node -e \"require('./src/app/startup').getApplicationHealth().then(r => console.log(JSON.stringify(r, null, 2)))\"",
    "api-report": "node -e \"require('./src/utils/apiDocumentation').generateValidationReport().then(r => console.log(r))\""
  }
}
```

### Manual Validation Steps

1. **Environment Variables Check**
   ```bash
   # Verify all required variables are set
   echo "API URL: $NEXT_PUBLIC_API_URL"
   echo "Environment: $NODE_ENV"
   echo "Monitoring: $MONITORING_ENABLED"
   ```

2. **API Connectivity Test**
   ```bash
   # Test API server connectivity
   curl -I $NEXT_PUBLIC_API_URL/api/v1/health
   ```

3. **Build Verification**
   ```bash
   # Ensure production build works
   npm run build
   npm run start &
   curl -I http://localhost:3000
   ```

## 🚨 Critical API Issues Prevention

### 1. Endpoint Validation System
The application includes comprehensive endpoint validation that:
- Compares frontend endpoints with Swagger documentation
- Validates critical endpoints during startup
- Provides detailed error reporting
- Offers automatic fixes suggestions

### 2. Runtime Health Monitoring
Production monitoring includes:
- Continuous endpoint health checking
- Automatic error detection and reporting
- Performance metrics collection
- Real-time status monitoring

### 3. Error Handling & Recovery
Robust error handling with:
- Retry logic with exponential backoff
- Graceful degradation for failed endpoints
- Comprehensive error logging
- User-friendly error messages

## 🏗️ Deployment Platforms

### Vercel Deployment

1. **Environment Variables Setup**
   ```bash
   # Set environment variables in Vercel dashboard
   vercel env add NEXT_PUBLIC_API_URL production
   vercel env add API_TIMEOUT production
   vercel env add MONITORING_ENABLED production
   # ... add all required variables
   ```

2. **Build Configuration**
   ```json
   // vercel.json
   {
     "build": {
       "env": {
         "NODE_ENV": "production"
       }
     },
     "env": {
       "NEXT_PUBLIC_API_URL": "@api-url-production"
     }
   }
   ```

### Netlify Deployment

1. **Build Settings**
   ```toml
   # netlify.toml
   [build]
     command = "npm run build"
     publish = "out"
   
   [build.environment]
     NODE_ENV = "production"
     NEXT_PUBLIC_API_URL = "https://api.yourdomain.com"
   ```

### Docker Deployment

1. **Production Dockerfile**
   ```dockerfile
   FROM node:18-alpine AS deps
   WORKDIR /app
   COPY package*.json ./
   RUN npm ci --only=production && npm cache clean --force
   
   FROM node:18-alpine AS builder
   WORKDIR /app
   COPY . .
   COPY --from=deps /app/node_modules ./node_modules
   RUN npm run build
   
   FROM node:18-alpine AS runner
   WORKDIR /app
   ENV NODE_ENV production
   COPY --from=builder /app/next.config.js ./
   COPY --from=builder /app/public ./public
   COPY --from=builder /app/.next ./.next
   COPY --from=builder /app/node_modules ./node_modules
   COPY --from=builder /app/package.json ./package.json
   
   EXPOSE 3000
   CMD ["npm", "start"]
   ```

## 📈 Monitoring & Maintenance

### 1. Health Check Endpoints

Create health check endpoints for monitoring:

```typescript
// pages/api/health.ts
import { getApplicationHealth } from '@/app/startup';

export default async function handler(req, res) {
  const health = await getApplicationHealth();
  res.status(health.status === 'healthy' ? 200 : 503).json(health);
}
```

### 2. Logging & Error Tracking

Configure structured logging:

```typescript
// utils/logger.ts
export function logAPIError(endpoint: string, error: Error, context?: any) {
  if (process.env.NODE_ENV === 'production') {
    // Send to monitoring service (Sentry, New Relic, etc.)
    console.error('API Error:', {
      endpoint,
      error: error.message,
      stack: error.stack,
      context,
      timestamp: new Date().toISOString()
    });
  }
}
```

### 3. Performance Monitoring

Monitor key metrics:
- API response times
- Error rates
- Cache hit rates
- User experience metrics

## 🚨 Troubleshooting

### Common Issues & Solutions

1. **404 API Errors**
   - Check `NEXT_PUBLIC_API_URL` configuration
   - Verify Next.js rewrites in `next.config.ts`
   - Ensure backend server is running and accessible

2. **Environment Variable Issues**
   - Run `npm run validate:env` to check configuration
   - Verify variables are properly set in deployment platform
   - Check variable names match exactly (case-sensitive)

3. **Endpoint Validation Failures**
   - Run `npm run validate:api` for detailed error report
   - Compare with Swagger documentation
   - Update `API_ENDPOINTS` configuration if needed

4. **Build Failures**
   - Check TypeScript compilation errors
   - Verify all dependencies are properly installed
   - Review build logs for specific error messages

### Emergency Rollback Procedure

1. **Immediate Actions**
   ```bash
   # Rollback to previous deployment
   vercel --prod rollback
   # or
   git revert HEAD && git push origin main
   ```

2. **Investigation Steps**
   - Check application logs
   - Review error monitoring dashboard
   - Run health checks on rolled-back version
   - Identify root cause before re-deploying

## 📞 Support Contacts

- **Development Team**: [team@yourdomain.com]
- **Infrastructure**: [infra@yourdomain.com]
- **Emergency Contact**: [emergency@yourdomain.com]

---

## 🎯 Success Criteria

Your application is ready for production when:

- [ ] ✅ All validation checks pass (>95% success rate)
- [ ] ✅ Production readiness score >90%
- [ ] ✅ Critical API endpoints are healthy
- [ ] ✅ Environment configuration is validated
- [ ] ✅ Security measures are properly configured
- [ ] ✅ Monitoring and error tracking are active
- [ ] ✅ Performance optimizations are enabled
- [ ] ✅ Build and deployment process is tested
- [ ] ✅ Rollback procedure is documented and tested

**Remember**: Never deploy to production with failing validation checks or unresolved critical issues!

---

*This guide is automatically updated with each release. Last updated: $(new Date().toISOString())*