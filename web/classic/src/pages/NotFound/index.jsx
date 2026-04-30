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

import React from 'react';
import { Button } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { Home } from 'lucide-react';

const NotFound = () => {
  const { t } = useTranslation();

  return (
    <div className='flex flex-col justify-center items-center min-h-screen bg-white dark:bg-gray-900 px-4'>
      {/* 404 数字 */}
      <h1
        className='text-9xl md:text-[180px] font-bold leading-none mb-6'
        style={{
          background: 'linear-gradient(135deg, #3B82F6 0%, #10B981 50%, #8B5CF6 100%)',
          WebkitBackgroundClip: 'text',
          WebkitTextFillColor: 'transparent',
          backgroundClip: 'text',
        }}
      >
        404
      </h1>

      {/* 提示标题 */}
      <h2 className='text-2xl md:text-3xl font-bold text-gray-900 dark:text-white mb-3'>
        页面未找到
      </h2>

      {/* 提示描述 */}
      <p className='text-base text-gray-500 dark:text-gray-400 mb-10 text-center max-w-md'>
        {t('页面未找到，请检查您的浏览器地址是否正确')}
      </p>

      {/* 返回首页按钮 */}
      <Link to='/'>
        <Button
          theme='solid'
          className='!rounded-full text-lg font-medium shadow-lg hover:shadow-xl transition-shadow'
          style={{
            background: 'linear-gradient(90deg, #3B82F6 0%, #10B981 100%)',
            border: 'none',
            width: 200,
            height: 48,
          }}
          icon={<Home size={20} />}
        >
          返回首页
        </Button>
      </Link>
    </div>
  );
};

export default NotFound;