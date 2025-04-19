import React from 'react';

const SimpleApp: React.FC = () => {
  return (
    <div style={{ 
      display: 'flex', 
      justifyContent: 'center', 
      alignItems: 'center', 
      height: '100vh',
      flexDirection: 'column',
      fontFamily: 'Arial, sans-serif'
    }}>
      <h1>中间人OpenAI API服务器</h1>
      <p>这是一个测试页面</p>
    </div>
  );
}

export default SimpleApp; 