import React from 'react';
import { Layout as AntLayout, theme } from 'antd';
import Navbar from '../Navbar';
import Footer from '../Footer';

const { Content, Footer: AntFooter } = AntLayout;

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
    <AntLayout>
      <Navbar />
      <Content style={{ 
        padding: '0 48px', 
        marginTop: 64, // 导航栏高度
        minHeight: 'calc(100vh - 64px)' // 减去导航栏高度
      }}>
        <div 
          style={{ 
            padding: 24, 
            marginTop: 24,
            background: colorBgContainer, 
            borderRadius: borderRadiusLG,
            minHeight: 'calc(100vh - 200px)'
          }}
        >
          {children}
        </div>
      </Content>
      <AntFooter style={{ 
        padding: 0,
        backgroundColor: 'transparent'
      }}>
        <Footer />
      </AntFooter>
    </AntLayout>
  );
};

export default Layout; 