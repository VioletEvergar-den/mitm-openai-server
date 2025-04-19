import React from 'react';
import { Layout, Typography } from 'antd';
import './Footer.css';

const { Footer: AntFooter } = Layout;
const { Link } = Typography;

const Footer: React.FC = () => {
  return (
    <AntFooter className="footer">
      <Link 
        href="https://github.com/llm-sec/mitm-openai-server" 
        target="_blank" 
        rel="noopener noreferrer"
      >
        MITM OpenAI Server
      </Link> © 2025
    </AntFooter>
  );
};

export default Footer; 