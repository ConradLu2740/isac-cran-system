#!/bin/bash
# ISAC-CRAN系统本地Minikube部署脚本
# 使用方法: ./scripts/deploy-minikube.sh [start|stop|status|logs]

set -e

NAMESPACE="isac-system"
APP_NAME="isac-cran-system"

print_banner() {
    echo "========================================"
    echo "  ISAC-CRAN系统 Kubernetes部署工具"
    echo "========================================"
}

check_minikube() {
    if ! command -v minikube &> /dev/null; then
        echo "❌ Minikube未安装，请先安装Minikube"
        echo "   参考: https://minikube.sigs.k8s.io/docs/start/"
        exit 1
    fi
    
    if ! command -v kubectl &> /dev/null; then
        echo "❌ kubectl未安装，请先安装kubectl"
        exit 1
    fi
    
    echo "✅ Minikube和kubectl已安装"
}

start_minikube() {
    echo "🚀 启动Minikube集群..."
    
    if minikube status &> /dev/null; then
        echo "✅ Minikube已在运行"
    else
        minikube start \
            --cpus=4 \
            --memory=8192 \
            --driver=docker \
            --kubernetes-version=v1.28.0
        echo "✅ Minikube启动成功"
    fi
    
    minikube addons enable metrics-server
    minikube addons enable ingress
    echo "✅ 必要插件已启用"
}

build_image() {
    echo "🔨 构建Docker镜像..."
    
    eval $(minikube docker-env)
    
    docker build -t isac-cran-system:latest .
    
    echo "✅ Docker镜像构建成功"
}

deploy_infrastructure() {
    echo "📦 部署基础设施组件..."
    
    kubectl apply -f - <<EOF
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
EOF
    
    echo "✅ 基础设施部署完成"
}

deploy_app() {
    echo "🚀 部署ISAC-CRAN应用..."
    
    kubectl apply -f k8s/deployment.yaml
    
    echo "⏳ 等待应用启动..."
    kubectl rollout status deployment/isac-server -n $NAMESPACE --timeout=120s
    
    echo "✅ 应用部署完成"
}

wait_for_infrastructure() {
    echo "⏳ 等待基础设施就绪..."
    
    kubectl wait --for=condition=ready pod -l app=mysql -n $NAMESPACE --timeout=120s || true
    kubectl wait --for=condition=ready pod -l app=redis -n $NAMESPACE --timeout=60s || true
    kubectl wait --for=condition=ready pod -l app=influxdb -n $NAMESPACE --timeout=120s || true
    
    sleep 5
    
    echo "✅ 基础设施就绪"
}

show_status() {
    echo ""
    echo "📊 部署状态:"
    echo "----------------------------------------"
    kubectl get pods -n $NAMESPACE
    echo ""
    echo "📊 服务状态:"
    kubectl get svc -n $NAMESPACE
    echo ""
    echo "📊 HPA状态:"
    kubectl get hpa -n $NAMESPACE 2>/dev/null || echo "HPA未配置"
    echo ""
    
    echo "🌐 访问方式:"
    echo "----------------------------------------"
    echo "  API地址: http://$(minikube ip):$(kubectl get svc isac-server -n $NAMESPACE -o jsonpath='{.spec.ports[0].nodePort}')"
    echo "  或使用: minikube service isac-server -n $NAMESPACE --url"
    echo ""
    echo "  pprof性能分析: http://<API地址>/debug/pprof/"
    echo "  运行时指标: http://<API地址>/debug/metrics"
}

show_logs() {
    echo "📋 应用日志 (Ctrl+C退出):"
    kubectl logs -f -l app=isac-server -n $NAMESPACE --tail=100
}

stop_deployment() {
    echo "🛑 停止部署..."
    kubectl delete -f k8s/deployment.yaml --ignore-not-found
    kubectl delete namespace $NAMESPACE --ignore-not-found
    echo "✅ 部署已停止"
}

port_forward() {
    echo "🔗 设置端口转发..."
    kubectl port-forward svc/isac-server -n $NAMESPACE 8080:8080 &
    echo "✅ 端口转发已设置: http://localhost:8080"
}

run_health_check() {
    echo "🏥 健康检查..."
    
    POD_NAME=$(kubectl get pods -n $NAMESPACE -l app=isac-server -o jsonpath='{.items[0].metadata.name}')
    
    kubectl exec -n $NAMESPACE $POD_NAME -- curl -s http://localhost:8080/api/v1/health || echo "健康检查失败"
    
    echo ""
}

case "${1:-start}" in
    start)
        print_banner
        check_minikube
        start_minikube
        build_image
        
        kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
        
        deploy_infrastructure
        wait_for_infrastructure
        deploy_app
        show_status
        ;;
    stop)
        print_banner
        stop_deployment
        ;;
    status)
        print_banner
        show_status
        ;;
    logs)
        show_logs
        ;;
    port-forward)
        port_forward
        ;;
    health)
        run_health_check
        ;;
    *)
        echo "使用方法: $0 {start|stop|status|logs|port-forward|health}"
        exit 1
        ;;
esac
