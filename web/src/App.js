import React, { useState, useEffect } from 'react';
import { Layout, Menu, Card, Row, Col, Statistic, Table, Button, message, Tabs, Descriptions, Divider } from 'antd';
import {
  DashboardOutlined,
  ApiOutlined,
  ExperimentOutlined,
  CloudServerOutlined,
  SettingOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import axios from 'axios';
import dayjs from 'dayjs';

const { Header, Sider, Content } = Layout;
const { TabPane } = Tabs;

const API_BASE = 'http://localhost:8080/api/v1';

function App() {
  const [collapsed, setCollapsed] = useState(false);
  const [activeMenu, setActiveMenu] = useState('dashboard');
  const [systemInfo, setSystemInfo] = useState(null);
  const [irsStatus, setIrsStatus] = useState(null);
  const [sensors, setSensors] = useState([]);
  const [loading, setLoading] = useState(false);
  const [beamformingResult, setBeamformingResult] = useState(null);
  const [doaResult, setDoaResult] = useState(null);

  useEffect(() => {
    fetchSystemInfo();
    fetchIRSStatus();
    fetchSensors();
  }, []);

  const fetchSystemInfo = async () => {
    try {
      const res = await axios.get(`${API_BASE}/info`);
      setSystemInfo(res.data.data);
    } catch (error) {
      console.error('Failed to fetch system info:', error);
    }
  };

  const fetchIRSStatus = async () => {
    try {
      const res = await axios.get(`${API_BASE}/irs/status`);
      setIrsStatus(res.data.data);
    } catch (error) {
      console.error('Failed to fetch IRS status:', error);
    }
  };

  const fetchSensors = async () => {
    try {
      const res = await axios.get(`${API_BASE}/sensor/list`);
      setSensors(res.data.data || []);
    } catch (error) {
      console.error('Failed to fetch sensors:', error);
    }
  };

  const runBeamforming = async () => {
    setLoading(true);
    try {
      const res = await axios.post(`${API_BASE}/algorithm/beamforming`, {
        experiment_id: `exp_${Date.now()}`,
        params: {
          element_count: 64,
          target_direction: 0.5,
          interference_angles: [0.2, 0.8],
          snr_threshold: 0.9,
          max_iterations: 100,
        },
      });
      message.success('波束成形算法运行成功！');
      setBeamformingResult(res.data.data);
    } catch (error) {
      message.error('算法运行失败');
      console.error(error);
    }
    setLoading(false);
  };

  const runDOA = async () => {
    setLoading(true);
    try {
      const res = await axios.post(`${API_BASE}/algorithm/doa`, {
        experiment_id: `exp_${Date.now()}`,
        params: {
          element_count: 64,
          num_sources: 3,
          snapshot_length: 1024,
          method: 'MUSIC',
        },
      });
      message.success('DOA估计算法运行成功！');
      setDoaResult(res.data.data);
    } catch (error) {
      message.error('算法运行失败');
      console.error(error);
    }
    setLoading(false);
  };

  const getBeamPatternOption = (result) => {
    if (!result || !result.beam_pattern) {
      const angles = [];
      const pattern = [];
      for (let i = 0; i < 360; i++) {
        angles.push(i);
        pattern.push(Math.sin(i * 0.1) * 0.5 + 0.5 + Math.random() * 0.1);
      }
      return {
        title: { text: '波束方向图', left: 'center' },
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: angles, name: '角度 (°)' },
        yAxis: { type: 'value', name: '增益' },
        series: [{ name: '增益', type: 'line', data: pattern, smooth: true, areaStyle: { opacity: 0.3 } }],
      };
    }

    const angles = [];
    const step = Math.max(1, Math.floor(result.beam_pattern.length / 360));
    for (let i = 0; i < result.beam_pattern.length; i += step) {
      angles.push(Math.floor(i * 360 / result.beam_pattern.length));
    }
    const sampledPattern = result.beam_pattern.filter((_, i) => i % step === 0);

    return {
      title: { text: '波束方向图 (算法结果)', left: 'center' },
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: angles, name: '角度 (°)' },
      yAxis: { type: 'value', name: '增益' },
      series: [{ name: '增益', type: 'line', data: sampledPattern, smooth: true, areaStyle: { opacity: 0.3 } }],
    };
  };

  const getDOASpectrumOption = (result) => {
    if (!result || !result.spectrum) {
      return {
        title: { text: 'DOA空间谱', left: 'center' },
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: [], name: '角度 (°)' },
        yAxis: { type: 'value', name: '谱值' },
        series: [{ name: '谱值', type: 'line', data: [], smooth: true }],
      };
    }

    const angles = [];
    const step = Math.max(1, Math.floor(result.spectrum.length / 180));
    for (let i = 0; i < result.spectrum.length; i += step) {
      angles.push((i * 180 / result.spectrum.length).toFixed(1));
    }
    const sampledSpectrum = result.spectrum.filter((_, i) => i % step === 0);

    return {
      title: { text: 'DOA空间谱 (MUSIC算法)', left: 'center' },
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: angles, name: '角度 (°)' },
      yAxis: { type: 'value', name: '谱值' },
      series: [{ name: '谱值', type: 'line', data: sampledSpectrum, smooth: true }],
    };
  };

  const getSensorChartOption = () => {
    return {
      title: { text: '传感器数据趋势', left: 'center' },
      tooltip: { trigger: 'axis' },
      legend: { data: ['温度', '湿度', '电压'], top: 30 },
      xAxis: { type: 'category', data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', '24:00'] },
      yAxis: { type: 'value' },
      series: [
        { name: '温度', type: 'line', data: [25, 26, 28, 30, 29, 27, 25] },
        { name: '湿度', type: 'line', data: [60, 58, 55, 50, 52, 56, 59] },
        { name: '电压', type: 'line', data: [220, 221, 219, 218, 220, 222, 221] },
      ],
    };
  };

  const menuItems = [
    { key: 'dashboard', icon: <DashboardOutlined />, label: '系统概览' },
    { key: 'irs', icon: <ApiOutlined />, label: 'IRS控制' },
    { key: 'algorithm', icon: <ExperimentOutlined />, label: '算法引擎' },
    { key: 'sensor', icon: <CloudServerOutlined />, label: '传感器' },
    { key: 'settings', icon: <SettingOutlined />, label: '系统设置' },
  ];

  const renderDashboard = () => (
    <div>
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card><Statistic title="IRS阵元数量" value={irsStatus?.element_count || 64} suffix="个" /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="工作频率" value="2.4" suffix="GHz" /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="传感器数量" value={sensors.length || 8} suffix="个" /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="系统状态" value="正常运行" valueStyle={{ color: '#3f8600' }} /></Card>
        </Col>
      </Row>
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="波束方向图">
            <ReactECharts option={getBeamPatternOption(beamformingResult)} style={{ height: 300 }} />
          </Card>
        </Col>
        <Col span={12}>
          <Card title="传感器数据趋势">
            <ReactECharts option={getSensorChartOption()} style={{ height: 300 }} />
          </Card>
        </Col>
      </Row>
    </div>
  );

  const renderIRSControl = () => (
    <Card title="IRS控制面板" extra={<Button icon={<ReloadOutlined />} onClick={fetchIRSStatus}>刷新状态</Button>}>
      <Row gutter={[16, 16]}>
        <Col span={12}><Statistic title="阵元数量" value={irsStatus?.element_count || 64} /></Col>
        <Col span={12}><Statistic title="频段" value={irsStatus?.frequency_band || '2.4GHz'} /></Col>
        <Col span={12}><Statistic title="温度" value={irsStatus?.temperature?.toFixed(1) || '29.4'} suffix="°C" /></Col>
        <Col span={12}><Statistic title="电源状态" value={irsStatus?.power_status ? '正常' : '异常'} valueStyle={{ color: irsStatus?.power_status ? '#3f8600' : '#cf1322' }} /></Col>
      </Row>
      <div style={{ marginTop: 24 }}>
        <h4>相移配置 (前16个)</h4>
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
          {irsStatus?.phase_shifts?.slice(0, 16).map((ps, i) => (
            <span key={i} style={{ padding: '2px 6px', background: '#f0f0f0', borderRadius: 4, fontSize: 12 }}>
              {ps.toFixed(2)}
            </span>
          ))}
        </div>
      </div>
    </Card>
  );

  const renderAlgorithm = () => (
    <Card title="算法引擎">
      <Tabs defaultActiveKey="beamforming">
        <TabPane tab="波束成形" key="beamforming">
          <div style={{ padding: 20 }}>
            <p>波束成形优化算法：通过调整IRS阵元的相移，实现信号在目标方向的增益最大化。</p>
            <Button type="primary" onClick={runBeamforming} loading={loading}>
              运行波束成形算法
            </Button>
            {beamformingResult && (
              <>
                <Divider>算法结果</Divider>
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <ReactECharts option={getBeamPatternOption(beamformingResult)} style={{ height: 250 }} />
                  </Col>
                  <Col span={12}>
                    <Descriptions column={2} bordered size="small">
                      <Descriptions.Item label="迭代次数">{beamformingResult.iterations}</Descriptions.Item>
                      <Descriptions.Item label="是否收敛">{beamformingResult.converged ? '是' : '否'}</Descriptions.Item>
                      <Descriptions.Item label="主瓣方向">{beamformingResult.main_lobe_direction?.toFixed(4)}</Descriptions.Item>
                      <Descriptions.Item label="副瓣电平">{beamformingResult.side_lobe_level?.toFixed(4)} dB</Descriptions.Item>
                    </Descriptions>
                  </Col>
                </Row>
              </>
            )}
          </div>
        </TabPane>
        <TabPane tab="DOA估计" key="doa">
          <div style={{ padding: 20 }}>
            <p>DOA（到达方向）估计算法：使用MUSIC算法估计信号源的到达方向。</p>
            <Button type="primary" onClick={runDOA} loading={loading}>
              运行DOA估计算法
            </Button>
            {doaResult && (
              <>
                <Divider>算法结果</Divider>
                <Row gutter={[16, 16]}>
                  <Col span={12}>
                    <ReactECharts option={getDOASpectrumOption(doaResult)} style={{ height: 250 }} />
                  </Col>
                  <Col span={12}>
                    <Descriptions column={1} bordered size="small">
                    <Descriptions.Item label="谱点数量">{doaResult.spectrum?.length}</Descriptions.Item>
                    </Descriptions>
                  </Col>
                </Row>
              </>
            )}
          </div>
        </TabPane>
      </Tabs>
    </Card>
  );

  const renderSensor = () => (
    <Card title="传感器列表">
      <Table
        dataSource={sensors}
        columns={[
          { title: '传感器ID', dataIndex: 'sensor_id', key: 'sensor_id' },
          { title: '类型', dataIndex: 'sensor_type', key: 'sensor_type' },
          { title: '位置', dataIndex: 'location', key: 'location' },
          { title: '单位', dataIndex: 'unit', key: 'unit' },
          { title: '状态', dataIndex: 'status', key: 'status', render: (s) => s === 1 ? '在线' : '离线' },
        ]}
        rowKey="sensor_id"
      />
    </Card>
  );

  const renderSettings = () => (
    <Card title="系统设置">
      <p>系统版本：{systemInfo?.version || '1.0.0'}</p>
      <p>系统名称：{systemInfo?.name || 'ISAC-CRAN System'}</p>
      <p>描述：{systemInfo?.description || 'ISAC-CRAN System'}</p>
    </Card>
  );

  const renderContent = () => {
    switch (activeMenu) {
      case 'dashboard': return renderDashboard();
      case 'irs': return renderIRSControl();
      case 'algorithm': return renderAlgorithm();
      case 'sensor': return renderSensor();
      case 'settings': return renderSettings();
      default: return renderDashboard();
    }
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}>
        <div style={{ height: 32, margin: 16, background: 'rgba(255, 255, 255, 0.2)', borderRadius: 6, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#fff', fontWeight: 'bold' }}>
          {collapsed ? 'ISAC' : 'ISAC-CRAN'}
        </div>
        <Menu theme="dark" selectedKeys={[activeMenu]} mode="inline" items={menuItems} onClick={(e) => setActiveMenu(e.key)} />
      </Sider>
      <Layout>
        <Header style={{ padding: '0 24px', background: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <h2 style={{ margin: 0 }}>ISAC-CRAN 实验原型系统</h2>
          <span style={{ color: '#666' }}>{dayjs().format('YYYY-MM-DD HH:mm:ss')}</span>
        </Header>
        <Content style={{ margin: '24px 16px', padding: 24, background: '#fff', minHeight: 280, borderRadius: 8 }}>
          {renderContent()}
        </Content>
      </Layout>
    </Layout>
  );
}

export default App;
