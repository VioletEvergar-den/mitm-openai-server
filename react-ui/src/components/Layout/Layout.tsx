import React from 'react';
import Navbar from '../Navbar/Navbar';
import './Layout.css';

const Layout: React.FC<{
  children: React.ReactNode;
  title?: string;
}> = ({ children, title }) => {
  // 设置页面标题
  React.useEffect(() => {
    if (title) {
      document.title = `${title} - 中间人OpenAI API服务器`;
    } else {
      document.title = '中间人OpenAI API服务器';
    }
  }, [title]);

  return (
    <div className="layout">
      <Navbar />
      <div className="layout-content container">
        {children}
      </div>
    </div>
  );
};

export default Layout; 