#!/usr/bin/env python3
"""
Script to remove unused API endpoints from Swagger documentation.
Based on analysis of frontend usage patterns.
"""

import yaml
import json
from pathlib import Path
import shutil
from datetime import datetime

def remove_unused_endpoints():
    """Remove unused API endpoints from swagger.yaml"""
    
    swagger_file = Path("D:/Project/app_sistem_akuntansi/backend/docs/swagger.yaml")
    backup_file = Path(f"D:/Project/app_sistem_akuntansi/backend/docs/swagger_backup_{datetime.now().strftime('%Y%m%d_%H%M%S')}.yaml")
    
    # Create backup
    print(f"Creating backup at: {backup_file}")
    shutil.copy2(swagger_file, backup_file)
    
    # List of unused endpoints to remove
    unused_endpoints = [
        # Journal Entry Auto-generation
        "/journal-entries/auto-generate/purchase",
        "/journal-entries/auto-generate/sale",
        
        # Journal Entry Operations
        "/journal-entries/{id}/post",
        "/journal-entries/{id}/reverse", 
        "/journal-entries/summary",
        
        # Account Journal Entries
        "/accounts/{account_id}/journal-entries",
        
        # Admin Operations
        "/api/admin/check-cashbank-gl-links",
        "/api/admin/fix-cashbank-gl-links",
        
        # Balance Monitoring (All endpoints)
        "/api/monitoring/balance-health",
        "/api/monitoring/balance-sync",
        "/api/monitoring/discrepancies", 
        "/api/monitoring/fix-discrepancies",
        "/api/monitoring/sync-status",
        
        # Payment Debug/Analytics
        "/api/payments/debug/recent",
        "/api/payments/analytics",
        "/api/payments/export/excel",
        "/api/payments/report/pdf",
        "/api/payments/{id}/pdf",
        
        # Enhanced Reports (All endpoints)
        "/api/reports/enhanced/financial-metrics",
        "/api/reports/enhanced/profit-loss",
        "/api/reports/enhanced/profit-loss-comparison",
        
        # Security Dashboard (All endpoints)
        "/api/v1/admin/security/alerts",
        "/api/v1/admin/security/alerts/{id}/acknowledge",
        "/api/v1/admin/security/cleanup",
        "/api/v1/admin/security/config",
        "/api/v1/admin/security/incidents",
        "/api/v1/admin/security/incidents/{id}",
        "/api/v1/admin/security/incidents/{id}/resolve",
        "/api/v1/admin/security/ip-whitelist",
        "/api/v1/admin/security/metrics",
    ]
    
    try:
        # Load YAML content
        print("Loading swagger.yaml...")
        with open(swagger_file, 'r', encoding='utf-8') as f:
            swagger_data = yaml.safe_load(f)
        
        # Remove unused endpoints
        removed_count = 0
        if 'paths' in swagger_data:
            original_count = len(swagger_data['paths'])
            
            for endpoint in unused_endpoints:
                if endpoint in swagger_data['paths']:
                    print(f"Removing unused endpoint: {endpoint}")
                    del swagger_data['paths'][endpoint]
                    removed_count += 1
                else:
                    print(f"Endpoint not found: {endpoint}")
            
            new_count = len(swagger_data['paths'])
            print(f"Removed {removed_count} unused endpoints ({original_count} -> {new_count})")
        
        # Add comment about the cleanup
        if 'info' not in swagger_data:
            swagger_data['info'] = {}
        
        original_description = swagger_data['info'].get('description', '')
        swagger_data['info']['description'] = f"{original_description}\n\nNOTE: Unused endpoints have been removed based on frontend usage analysis performed on {datetime.now().strftime('%Y-%m-%d')}."
        
        # Write updated YAML
        print(f"Writing cleaned swagger.yaml...")
        with open(swagger_file, 'w', encoding='utf-8') as f:
            yaml.dump(swagger_data, f, default_flow_style=False, sort_keys=False, allow_unicode=True)
        
        print(f"✅ Successfully removed {removed_count} unused endpoints from Swagger documentation")
        print(f"✅ Backup saved to: {backup_file}")
        
        # Generate summary report
        generate_cleanup_report(unused_endpoints, removed_count, backup_file)
        
    except Exception as e:
        print(f"❌ Error processing swagger.yaml: {e}")
        # Restore backup if something went wrong
        if backup_file.exists():
            shutil.copy2(backup_file, swagger_file)
            print(f"🔄 Restored original file from backup")
        raise

def generate_cleanup_report(unused_endpoints, removed_count, backup_file):
    """Generate a report of the cleanup operation"""
    
    report_file = Path("D:/Project/app_sistem_akuntansi/SWAGGER_CLEANUP_REPORT.md")
    
    report_content = f"""# Swagger Cleanup Report

## Summary
- **Date**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
- **Total endpoints targeted for removal**: {len(unused_endpoints)}
- **Successfully removed**: {removed_count}
- **Backup location**: {backup_file}

## Removed Endpoints

### Journal Entry Operations
- `/journal-entries/auto-generate/purchase` (POST)
- `/journal-entries/auto-generate/sale` (POST)
- `/journal-entries/{{id}}/post` (POST)
- `/journal-entries/{{id}}/reverse` (POST)
- `/journal-entries/summary` (GET)

### Account Operations
- `/accounts/{{account_id}}/journal-entries` (GET)

### Admin Operations  
- `/api/admin/check-cashbank-gl-links` (GET)
- `/api/admin/fix-cashbank-gl-links` (POST)

### Balance Monitoring
- `/api/monitoring/balance-health` (GET)
- `/api/monitoring/balance-sync` (GET)
- `/api/monitoring/discrepancies` (GET)
- `/api/monitoring/fix-discrepancies` (POST)
- `/api/monitoring/sync-status` (GET)

### Payment Analytics & Export
- `/api/payments/debug/recent` (GET)
- `/api/payments/analytics` (GET)
- `/api/payments/export/excel` (GET)
- `/api/payments/report/pdf` (GET)
- `/api/payments/{{id}}/pdf` (GET)

### Enhanced Reports
- `/api/reports/enhanced/financial-metrics` (GET)
- `/api/reports/enhanced/profit-loss` (GET)
- `/api/reports/enhanced/profit-loss-comparison` (GET)

### Security Dashboard
- `/api/v1/admin/security/alerts` (GET)
- `/api/v1/admin/security/alerts/{{id}}/acknowledge` (PUT)
- `/api/v1/admin/security/cleanup` (POST)
- `/api/v1/admin/security/config` (GET)
- `/api/v1/admin/security/incidents` (GET)
- `/api/v1/admin/security/incidents/{{id}}` (GET)
- `/api/v1/admin/security/incidents/{{id}}/resolve` (PUT)
- `/api/v1/admin/security/ip-whitelist` (GET, POST)
- `/api/v1/admin/security/metrics` (GET)

## Notes
- All removed endpoints were identified as unused based on comprehensive frontend code analysis
- Backend implementation remains intact - only Swagger documentation was cleaned
- Some endpoints may be used by external integrations or admin tools not covered in this analysis
- To restore original Swagger file, use the backup located at: `{backup_file}`

## Next Steps
1. ✅ Swagger documentation cleaned
2. ⏳ Test Swagger UI to ensure it loads correctly
3. ⏳ Verify API documentation reflects only used endpoints
4. 📋 Consider implementing any useful endpoints that should be in frontend
"""

    with open(report_file, 'w', encoding='utf-8') as f:
        f.write(report_content)
    
    print(f"📋 Generated cleanup report: {report_file}")

if __name__ == "__main__":
    remove_unused_endpoints()