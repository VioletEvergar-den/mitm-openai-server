import React from 'react';
import { Layout, Typography } from 'antd';
import { GithubOutlined } from '@ant-design/icons';
import { useTheme } from '../../contexts/ThemeContext';
import './Footer.css';

const { Footer: AntFooter } = Layout;
const { Link, Text } = Typography;

const Footer: React.FC = () => {
  const { mode } = useTheme();
  
  return (
    <AntFooter className={`footer ${mode === 'dark' ? 'footer-dark' : ''}`}>
      <div className="footer-content">
        <Link 
          href="https://github.com/llm-sec/mitm-openai-server" 
          target="_blank" 
          rel="noopener noreferrer"
          className="footer-link"
        >
          <GithubOutlined style={{ marginRight: 6 }} />
          MITM OpenAI Server
        </Link>
        <span className="footer-divider">|</span>
        <Text className="footer-copyright">
          © 2025
        </Text>
      </div>
    </AntFooter>
  );
};

export default Footer;
