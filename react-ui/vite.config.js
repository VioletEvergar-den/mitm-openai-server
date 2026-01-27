import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { resolve } from 'path'

export default defineConfig({
  plugins: [react()],
  root: '.',
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    // 前端开发服务器端口修改为57691，避免与常用端口5173冲突
    // 注意：此端口为项目指定端口，严禁随意修改为其他端口
    port: 57691,
    strictPort: false,
    open: true,
    proxy: {
      '/ui/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
        ws: true,
        configure: (proxy, _options) => {
          proxy.on('error', (err, _req, _res) => {
            console.log('代理错误', err);
          });
          proxy.on('proxyReq', (proxyReq, req, _res) => {
            console.log('正在代理请求:', req.method, req.url);
            
            // 确保Authorization头被正确传递
            const authHeader = req.headers['authorization'];
            if (authHeader) {
              proxyReq.setHeader('Authorization', authHeader);
              console.log('已设置Authorization头:', authHeader);
            }
            
            // 尝试从X-Auth-Token自定义头获取
            const xAuthToken = req.headers['x-auth-token'];
            if (xAuthToken) {
              proxyReq.setHeader('Authorization', `Bearer ${xAuthToken}`);
              console.log('从X-Auth-Token设置了Authorization头');
            }
            
            // 尝试从cookie获取令牌
            if (req.headers.cookie) {
              const matches = req.headers.cookie.match(/auth_token=([^;]+)/);
              if (matches && matches[1]) {
                proxyReq.setHeader('Authorization', `Bearer ${matches[1]}`);
                console.log('从cookie设置了Authorization头');
              }
            }
            
            // 输出完整的请求头信息进行调试
            console.log('代理请求头:', proxyReq.getHeaders());
          });
          
          // 记录代理响应
          proxy.on('proxyRes', (proxyRes, req, _res) => {
            console.log(`代理响应: ${req.method} ${req.url} => ${proxyRes.statusCode}`);
            // 检查是否有401错误
            if (proxyRes.statusCode === 401) {
              console.error('收到401未授权响应!', {
                path: req.url,
                headers: req.headers
              });
            }
          });
        }
      },
      '/v1': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,
      }
    },
  },
  base: '/ui/',
  build: {
    outDir: 'dist',
    emptyOutDir: true
  }
})