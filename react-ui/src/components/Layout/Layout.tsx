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
    token: { colorBgContainer, borderRadiusLG, colorBgLayout },
  } = theme.useToken();

  React.useEffect(() => {
    if (title) {
      document.title = `${title} - OpenAI API 代理`;
    } else {
      document.title = 'OpenAI API 代理';
    }
  }, [title]);

  return (
    <AntLayout style={{ minHeight: '100vh', background: colorBgLayout }}>
      <Navbar />
      <Content style={{ 
        padding: '24px 48px', 
        marginTop: 64,
        minHeight: 'calc(100vh - 64px)'
      }}>
        <div 
          style={{ 
            padding: 24, 
            background: colorBgContainer, 
            borderRadius: borderRadiusLG,
            minHeight: 'calc(100vh - 180px)'
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
