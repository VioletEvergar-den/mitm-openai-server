import React from 'react';
import { Layout as AntLayout, theme } from 'antd';
import Navbar from '../Navbar';
import Footer from '../Footer';

const { Content } = AntLayout;

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
        padding: '24px 32px', 
        marginTop: 64,
        minHeight: 'calc(100vh - 64px)',
        maxWidth: '1600px',
        margin: '64px auto 0',
        width: '100%'
      }}>
        <div 
          style={{ 
            padding: 24, 
            background: colorBgContainer, 
            borderRadius: borderRadiusLG,
            minHeight: 'calc(100vh - 160px)',
            boxShadow: '0 1px 2px 0 rgba(0, 0, 0, 0.03), 0 1px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px 0 rgba(0, 0, 0, 0.02)'
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
