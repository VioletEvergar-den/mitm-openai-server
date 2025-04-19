import React from 'react';
import { Layout as AntLayout, theme } from 'antd';
import Navbar from '../Navbar/Navbar';
import Footer from '../Footer';
import './Layout.css';

const { Header, Content } = AntLayout;

const Layout: React.FC<{
  children: React.ReactNode;
  title?: string;
}> = ({ children, title }) => {
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  // 设置页面标题
  React.useEffect(() => {
    if (title) {
      document.title = `${title} - 中间人OpenAI API服务器`;
    } else {
      document.title = '中间人OpenAI API服务器';
    }
  }, [title]);

  return (
    <AntLayout className="layout" style={{ minHeight: '100vh' }}>
      <Header style={{ padding: 0, backgroundColor: '#001529' }}>
      <Navbar />
      </Header>
      <Content style={{ padding: '0 50px', marginTop: 16 }}>
        <div 
          style={{ 
            padding: 24, 
            background: colorBgContainer, 
            borderRadius: borderRadiusLG,
            minHeight: 280
          }}
        >
        {children}
      </div>
      </Content>
      <Footer />
    </AntLayout>
  );
};

export default Layout; 