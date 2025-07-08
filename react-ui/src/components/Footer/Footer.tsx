import React from 'react';
import { Layout } from 'antd';
import './Footer.css';

const { Footer: AntFooter } = Layout;

const Footer: React.FC = () => {
  return (
    <AntFooter className="footer">
      <div className="footer-content">
        <a 
          href="https://github.com/llm-sec/mitm-openai-server" 
          target="_blank" 
          rel="noopener noreferrer"
          className="footer-link"
        >
          MITM OpenAI Server
        </a>
        <span className="footer-divider">•</span>
        <span className="footer-copyright">2025</span>
      </div>
    </AntFooter>
  );
};

export default Footer; 