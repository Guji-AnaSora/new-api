/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useContext, useEffect, useState } from 'react';
import { Button, Typography } from '@douyinfe/semi-ui';
import { API, showError } from '../../helpers';
import { useIsMobile } from '../../hooks/common/useIsMobile';
import { StatusContext } from '../../context/Status';
import { useActualTheme } from '../../context/Theme';
import { marked } from 'marked';
import { useTranslation } from 'react-i18next';
import { IconApps } from '@douyinfe/semi-icons';
import { Globe, Shield, Zap, Sparkles } from 'lucide-react';
import { Link } from 'react-router-dom';
import NoticeModal from '../../components/layout/NoticeModal';

const { Text } = Typography;

const Home = () => {
  const { t, i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const actualTheme = useActualTheme();
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [noticeVisible, setNoticeVisible] = useState(false);
  const isMobile = useIsMobile();
  const isChinese = i18n.language?.startsWith('zh') ?? true;

  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);

      // 如果内容是 URL，则发送主题模式
      if (data.startsWith('https://')) {
        const iframe = document.querySelector('iframe');
        if (iframe) {
          iframe.onload = () => {
            iframe.contentWindow.postMessage({ themeMode: actualTheme }, '*');
            iframe.contentWindow.postMessage({ lang: i18n.language }, '*');
          };
        }
      }
    } else {
      showError(message);
      setHomePageContent('加载首页内容失败...');
    }
    setHomePageContentLoaded(true);
  };

  useEffect(() => {
    const checkNoticeAndShow = async () => {
      const lastCloseDate = localStorage.getItem('notice_close_date');
      const today = new Date().toDateString();
      if (lastCloseDate !== today) {
        try {
          const res = await API.get('/api/notice');
          const { success, data } = res.data;
          if (success && data && data.trim() !== '') {
            setNoticeVisible(true);
          }
        } catch (error) {
          console.error('获取公告失败:', error);
        }
      }
    };

    checkNoticeAndShow();
  }, []);

  useEffect(() => {
    displayHomePageContent().then();
  }, []);

  // 特性卡片数据
  const features = [
    {
      icon: <Globe size={56} className='text-[#bfdbfe]' fill='#bfdbfe' strokeWidth={1} />,
      title: '全网直连',
      subtitle: '翻墙不稳定、速度慢、门槛高？',
      description: 'CodeMax国内专线直连，无需任何第三方网络工具',
    },
    {
      icon: <Shield size={56} className='text-[#bbf7d0]' fill='#bbf7d0' strokeWidth={1} />,
      title: '极致稳定',
      subtitle: '账号高频被封，业务频繁中断？',
      description: 'CodeMax统一接入企业级API，告别封号噩梦，确保业务持续在线',
    },
    {
      icon: <IconApps size={56} className='text-[#ddd6fe]' />,
      title: '模型聚合',
      subtitle: '不同模型反复切换，Token浪费？',
      description: 'CodeMax一个账号集成Claude、GPT、DeepSeek等顶级模型，根据任务复杂度灵活切换',
    },
    {
      icon: <Zap size={56} className='text-[#fed7aa]' fill='#fed7aa' strokeWidth={1} />,
      title: '零感接入',
      subtitle: '安装繁琐，没技术底蕴配置不了？',
      description: 'CodeMax，无需任何配置，直接使用，让每个人都能零门槛享受AI红利',
    },
  ];

  return (
    <div className='w-full overflow-x-hidden'>
      <NoticeModal
        visible={noticeVisible}
        onClose={() => setNoticeVisible(false)}
        isMobile={isMobile}
      />
      {homePageContentLoaded && homePageContent === '' ? (
        <div className='w-full overflow-x-hidden min-h-screen bg-gradient-to-b from-blue-50/50 via-white to-white dark:from-gray-900/50 dark:via-gray-900 dark:to-gray-900'>
          {/* Banner 部分 */}
          <div className='w-full min-h-[600px] md:min-h-[700px] relative overflow-x-hidden flex flex-col items-center justify-center px-4 py-20'>
            {/* 背景装饰 */}
            <div className='absolute inset-0 overflow-hidden pointer-events-none'>
              <div className='absolute top-10 left-1/4 w-96 h-96 bg-blue-200/30 dark:bg-blue-500/10 rounded-full blur-3xl' />
              <div className='absolute bottom-10 right-1/4 w-80 h-80 bg-purple-200/30 dark:bg-purple-500/10 rounded-full blur-3xl' />
              <div className='absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[500px] h-[500px] bg-cyan-100/20 dark:bg-cyan-500/5 rounded-full blur-3xl' />
            </div>

            {/* 内容区 */}
            <div className='relative z-10 flex flex-col items-center text-center max-w-4xl mx-auto'>
              {/* CodeMax 品牌标题 */}
              <h1
                className={`text-5xl md:text-6xl lg:text-7xl font-bold mb-6 ${isChinese ? 'tracking-wide' : ''}`}
                style={{
                  background: 'linear-gradient(90deg, #3B82F6 0%, #10B981 25%, #F59E0B 50%, #8B5CF6 75%, #EC4899 100%)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                  backgroundClip: 'text',
                }}
              >
                CodeMax
              </h1>

              {/* 副标题 */}
              <h2
                className={`text-2xl md:text-3xl lg:text-4xl font-bold text-semi-color-text-0 mb-4 ${isChinese ? 'tracking-wide' : ''}`}
              >
                集全球智慧，让每一个诉求一键成真
              </h2>

              {/* 标签行 */}
              <div className='flex flex-wrap items-center justify-center gap-2 md:gap-4 text-sm md:text-base text-semi-color-text-1 mb-10'>
                <span>打破网络壁垒</span>
                <span className='text-semi-color-text-2'>·</span>
                <span>无需环境配置</span>
                <span className='text-semi-color-text-2'>·</span>
                <span>远离封号风险</span>
                <span className='text-semi-color-text-2'>·</span>
                <span>整合顶级模型</span>
              </div>

              {/* 主按钮 */}
              <Link to='/console'>
                <Button
                  theme='solid'
                  className='rounded-[25px] text-xl font-medium shadow-lg hover:shadow-xl transition-shadow'
                  style={{
                    background: 'linear-gradient(90deg, #3B82F6 0%, #10B981 100%)',
                    border: 'none',
                    width: 250,
                    height: 50,
                  }}
                  icon={<Sparkles size={18} />}
                >
                  开启AI体验
                </Button>
              </Link>
            </div>
          </div>

          {/* 特性卡片区域 */}
          <div className='w-full px-4 md:px-8 lg:px-16 pb-10 pt-[100px] bg-[#f6f9fb] dark:bg-gray-900/30'>
            <div className='max-w-6xl mx-auto grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4'>
              {features.map((feature, index) => (
                <div
                  key={index}
                  className='bg-white dark:bg-gray-800 rounded-xl p-4 shadow-[0_1px_8px_rgba(0,0,0,0.03)] dark:shadow-none'
                >
                  <div className='mb-3'>
                    {feature.icon}
                  </div>
                  <h3 className='text-lg font-bold text-gray-900 dark:text-white mb-1'>
                    {feature.title}
                  </h3>
                  <p className='text-sm text-gray-800 dark:text-gray-200 mb-2 leading-relaxed'>
                    {feature.subtitle}
                  </p>
                  <p className='text-xs text-gray-400 dark:text-gray-500 leading-relaxed'>
                    {feature.description}
                  </p>
                </div>
              ))}
            </div>
          </div>
        </div>
      ) : (
        <div className='overflow-x-hidden w-full'>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              className='w-full h-screen border-none'
            />
          ) : (
            <div
              className='mt-[60px]'
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            />
          )}
        </div>
      )}
    </div>
  );
};

export default Home;