import React from 'react';
import { Layout as AntLayout } from 'antd';
import Footer from './components/Footer';
import './styles/App.css';

const { Content } = AntLayout;

const TempApp: React.FC = () => {
  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Content style={{ padding: '50px 20px' }}>
        <div className="container">
          <div className="card" style={{ marginTop: '50px' }}>
            <h1 className="card-title">中间人OpenAI API服务器</h1>
            <p>应用正在加载中...</p>
            <p>这是一个临时页面，用于测试Vite开发服务器是否正常工作。</p>
          </div>
        </div>
      </Content>
      <Footer />
    </AntLayout>
  );
};

export default TempApp; 