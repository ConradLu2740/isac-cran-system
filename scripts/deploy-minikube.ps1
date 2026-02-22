# ISAC-CRAN系统本地Minikube部署脚本 (Windows PowerShell)
# 使用方法: .\scripts\deploy-minikube.ps1 [start|stop|status|logs]

param(
    [Parameter(Position=0)]
    [ValidateSet("start", "stop", "status", "logs", "port-forward", "health")]
    [string]$Action = "start"
)

$ErrorActionPreference = "Stop"
$Namespace = "isac-system"
$AppName = "isac-cran-system"

function Print-Banner {
    Write-Host "========================================"
    Write-Host "  ISAC-CRAN系统 Kubernetes部署工具"
    Write-Host "========================================"
}

function Check-Minikube {
    if (-not (Get-Command minikube -ErrorAction SilentlyContinue)) {
        Write-Host "❌ Minikube未安装，请先安装Minikube" -ForegroundColor Red
        Write-Host "   参考: https://minikube.sigs.k8s.io/docs/start/"
        exit 1
    }
    
    if (-not (Get-Command kubectl -ErrorAction SilentlyContinue)) {
        Write-Host "❌ kubectl未安装，请先安装kubectl" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "✅ Minikube和kubectl已安装" -ForegroundColor Green
}

function Start-Minikube {
    Write-Host "🚀 启动Minikube集群..." -ForegroundColor Cyan
    
    $status = minikube status 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Minikube已在运行" -ForegroundColor Green
    } else {
        minikube start --cpus=4 --memory=8192 --driver=docker --kubernetes-version=v1.28.0
        Write-Host "✅ Minikube启动成功" -ForegroundColor Green
    }
    
    minikube addons enable metrics-server
    minikube addons enable ingress
    Write-Host "✅ 必要插件已启用" -ForegroundColor Green
}

function Build-Image {
    Write-Host "🔨 构建Docker镜像..." -ForegroundColor Cyan
    
    minikube docker-env | Invoke-Expression
    
    docker build -t isac-cran-system:latest .
    
    Write-Host "✅ Docker镜像构建成功" -ForegroundColor Green
}

function Deploy-Infrastructure {
    Write-Host "📦 部署基础设施组件..." -ForegroundColor Cyan
    
    $infraYaml = @"
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: mysql-pvc
  namespace: isac-system
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: influxdb-pvc
  namespace: isac-system
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
  namespace: isac-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mysql
  template:
    metadata:
      labels:
        app: mysql
    spec:
      containers:
      - name: mysql
        image: mysql:8.0
        env:
        - name: MYSQL_ROOT_PASSWORD
          value: "root123"
        - name: MYSQL_DATABASE
          value: "isac_cran"
        ports:
        - containerPort: 3306
        volumeMounts:
        - name: mysql-storage
          mountPath: /var/lib/mysql
      volumes:
      - name: mysql-storage
        persistentVolumeClaim:
          claimName: mysql-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: mysql-service
  namespace: isac-system
spec:
  selector:
    app: mysql
  ports:
  - port: 3306
    targetPort: 3306
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: isac-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
---
apiVersion: v1
kind: Service
metadata:
  name: redis-service
  namespace: isac-system
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: influxdb
  namespace: isac-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: influxdb
  template:
    metadata:
      labels:
        app: influxdb
    spec:
      containers:
      - name: influxdb
        image: influxdb:2.7
        env:
        - name: DOCKER_INFLUXDB_INIT_MODE
          value: "setup"
        - name: DOCKER_INFLUXDB_INIT_USERNAME
          value: "admin"
        - name: DOCKER_INFLUXDB_INIT_PASSWORD
          value: "admin123"
        - name: DOCKER_INFLUXDB_INIT_ORG
          value: "isac-lab"
        - name: DOCKER_INFLUXDB_INIT_BUCKET
          value: "channel-data"
        - name: DOCKER_INFLUXDB_INIT_ADMIN_TOKEN
          value: "my-token"
        ports:
        - containerPort: 8086
        volumeMounts:
        - name: influxdb-storage
          mountPath: /var/lib/influxdb2
      volumes:
      - name: influxdb-storage
        persistentVolumeClaim:
          claimName: influxdb-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: influxdb-service
  namespace: isac-system
spec:
  selector:
    app: influxdb
  ports:
  - port: 8086
    targetPort: 8086
"@
    
    $infraYaml | kubectl apply -f -
    
    Write-Host "✅ 基础设施部署完成" -ForegroundColor Green
}

function Wait-Infrastructure {
    Write-Host "⏳ 等待基础设施就绪..." -ForegroundColor Cyan
    
    kubectl wait --for=condition=ready pod -l app=mysql -n $Namespace --timeout=120s 2>$null
    kubectl wait --for=condition=ready pod -l app=redis -n $Namespace --timeout=60s 2>$null
    kubectl wait --for=condition=ready pod -l app=influxdb -n $Namespace --timeout=120s 2>$null
    
    Start-Sleep -Seconds 5
    
    Write-Host "✅ 基础设施就绪" -ForegroundColor Green
}

function Deploy-App {
    Write-Host "🚀 部署ISAC-CRAN应用..." -ForegroundColor Cyan
    
    kubectl apply -f k8s/deployment.yaml
    
    Write-Host "⏳ 等待应用启动..." -ForegroundColor Cyan
    kubectl rollout status deployment/isac-server -n $Namespace --timeout=120s
    
    Write-Host "✅ 应用部署完成" -ForegroundColor Green
}

function Show-Status {
    Write-Host ""
    Write-Host "📊 部署状态:" -ForegroundColor Yellow
    Write-Host "----------------------------------------"
    kubectl get pods -n $Namespace
    Write-Host ""
    Write-Host "📊 服务状态:" -ForegroundColor Yellow
    kubectl get svc -n $Namespace
    Write-Host ""
    Write-Host "📊 HPA状态:" -ForegroundColor Yellow
    kubectl get hpa -n $Namespace 2>$null
    Write-Host ""
    
    Write-Host "🌐 访问方式:" -ForegroundColor Yellow
    Write-Host "----------------------------------------"
    $minikubeIp = minikube ip
    $nodePort = kubectl get svc isac-server -n $Namespace -o jsonpath='{.spec.ports[0].nodePort}'
    Write-Host "  API地址: http://${minikubeIp}:${nodePort}"
    Write-Host ""
    Write-Host "  pprof性能分析: http://${minikubeIp}:${nodePort}/debug/pprof/"
    Write-Host "  运行时指标: http://${minikubeIp}:${nodePort}/debug/metrics"
}

function Show-Logs {
    Write-Host "📋 应用日志 (Ctrl+C退出):" -ForegroundColor Cyan
    kubectl logs -f -l app=isac-server -n $Namespace --tail=100
}

function Stop-Deployment {
    Write-Host "🛑 停止部署..." -ForegroundColor Cyan
    kubectl delete -f k8s/deployment.yaml --ignore-not-found
    kubectl delete namespace $Namespace --ignore-not-found 2>$null
    Write-Host "✅ 部署已停止" -ForegroundColor Green
}

function Start-PortForward {
    Write-Host "🔗 设置端口转发..." -ForegroundColor Cyan
    Start-Job -ScriptBlock { kubectl port-forward svc/isac-server -n isac-system 8080:8080 }
    Write-Host "✅ 端口转发已设置: http://localhost:8080" -ForegroundColor Green
}

function Run-HealthCheck {
    Write-Host "🏥 健康检查..." -ForegroundColor Cyan
    
    $podName = kubectl get pods -n $Namespace -l app=isac-server -o jsonpath='{.items[0].metadata.name}'
    
    kubectl exec -n $Namespace $podName -- curl -s http://localhost:8080/api/v1/health
    
    Write-Host ""
}

# 主逻辑
switch ($Action) {
    "start" {
        Print-Banner
        Check-Minikube
        Start-Minikube
        Build-Image
        
        kubectl create namespace $Namespace --dry-run=client -o yaml | kubectl apply -f -
        
        Deploy-Infrastructure
        Wait-Infrastructure
        Deploy-App
        Show-Status
    }
    "stop" {
        Print-Banner
        Stop-Deployment
    }
    "status" {
        Print-Banner
        Show-Status
    }
    "logs" {
        Show-Logs
    }
    "port-forward" {
        Start-PortForward
    }
    "health" {
        Run-HealthCheck
    }
}
