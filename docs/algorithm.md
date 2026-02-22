# 算法说明文档

## 1. 波束成形优化算法

### 1.1 算法原理

波束成形通过调整天线阵列各阵元的加权系数，使信号在期望方向上形成主波束，同时在干扰方向上形成零陷。

### 1.2 数学模型

对于N元均匀线阵，阵列响应向量：

```
a(θ) = [1, e^(j2πd sinθ/λ), ..., e^(j2π(N-1)d sinθ/λ)]^T
```

波束成形输出：
```
y = w^H * x
```

其中 w 为权向量，x 为接收信号向量。

### 1.3 优化目标

最大化信干噪比（SINR）：
```
max w^H R_s w
s.t. w^H R_i+n w ≤ ε
     ||w||^2 = 1
```

### 1.4 实现说明

系统实现了共轭波束成形和MVDR波束成形两种方法：

- **共轭波束成形**：权向量取目标方向导向矢量的共轭
- **MVDR波束成形**：最小方差无失真响应，需要在干扰方向形成零陷

## 2. DOA估计算法

### 2.1 MUSIC算法

MUSIC（Multiple Signal Classification）是一种高分辨率的DOA估计算法。

#### 原理

1. 计算接收信号的协方差矩阵 R
2. 对R进行特征分解
3. 利用噪声子空间构建空间谱
4. 通过谱峰搜索估计DOA

#### 空间谱公式

```
P_MUSIC(θ) = 1 / (a^H(θ) * U_n * U_n^H * a(θ))
```

其中 U_n 为噪声子空间。

### 2.2 ESPRIT算法

ESPRIT利用阵列的平移不变性进行DOA估计，无需谱峰搜索。

## 3. 资源调度算法

### 3.1 调度策略

系统实现了三种调度策略：

| 策略 | 描述 |
|------|------|
| Round Robin | 轮询调度，公平分配资源 |
| Priority | 优先级调度，高优先级用户优先 |
| Proportional Fair | 比例公平，平衡吞吐量和公平性 |

### 3.2 实现细节

```go
type Scheduler struct {
    users      map[int]*User
    resources  []*Resource
    algorithm  SchedulingAlgorithm
}

func (s *Scheduler) Schedule() map[int]int {
    switch s.algorithm {
    case AlgorithmRoundRobin:
        return s.roundRobin()
    case AlgorithmPriority:
        return s.priorityBased()
    case AlgorithmProportionalFair:
        return s.proportionalFair()
    }
}
```

## 4. 无速率编码

### 4.1 LT码

Luby Transform (LT) 码是一种实用的无速率码。

#### 编码过程

1. 根据度分布函数采样度值 d
2. 随机选择 d 个源符号
3. 对选中的符号进行异或操作

#### 解码过程

采用BP（Belief Propagation）解码：

1. 找到度为1的编码符号
2. 恢复对应的源符号
3. 更新其他编码符号
4. 重复直到所有源符号恢复

### 4.2 鲁棒孤子分布

系统使用鲁棒孤子分布作为度分布函数：

```
ρ(d) = c * ln(K/δ) / (d * K)  for d = 1, ..., K/δ
τ(d) = 1/K                     for d = K/δ
```

## 5. MATLAB协同

### 5.1 数据交换格式

系统支持JSON和MAT文件两种格式：

- **JSON格式**：轻量级，适合小数据量
- **MAT文件**：MATLAB原生格式，支持复数和矩阵

### 5.2 导出接口

```go
func (s *MATLABService) ExportToJSON(data interface{}, filename string) error
func (s *MATLABService) ExportToCSV(data interface{}, filename string) error
```

### 5.3 MATLAB脚本示例

```matlab
% beamforming_validation.m
% 从JSON文件加载波束成形结果
data = jsondecode(fileread('beamforming_result.json'));

% 绘制波束方向图
figure;
polarplot(data.beam_pattern);
title('Beam Pattern Validation');
```
