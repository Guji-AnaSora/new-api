import React, { useEffect, useState } from 'react';
import {
  Layout,
  Nav,
  Card,
  Typography,
  Tag,
  Button,
  Empty,
  Spin,
  Input,
  Banner,
  Modal,
  Descriptions,
  Divider,
  Toast,
  SideSheet,
  Avatar,
} from '@douyinfe/semi-ui';
import {
  IconSearch,
  IconDownload,
  IconGlobeStroke,
  IconCalendarClock,
  IconInfoCircle,
  IconChevronLeft,
  IconFile,
  IconWrench,
  IconUser,
} from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import { useIsMobile } from '../../hooks/common/useIsMobile';

const { Title, Text, Paragraph } = Typography;
const { Content } = Layout;

const SkillMarketplace = () => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const [skills, setSkills] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [activeSearchQuery, setActiveSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [detailVisible, setDetailVisible] = useState(false);
  const [selectedSkill, setSelectedSkill] = useState(null);
  const [downloadingId, setDownloadingId] = useState(null);
  const [dynamicCategories, setDynamicCategories] = useState(null);
  const [docsLink, setDocsLink] = useState('');

  const defaultCategories = [
    { key: 'all', label: t('全部'), icon: <IconGlobeStroke /> },
  ];

  const categories = dynamicCategories || defaultCategories;

  useEffect(() => {
    fetchSkills();
    fetchCategories();
    fetchDocsLink();
  }, []);

  const fetchDocsLink = async () => {
    try {
      const res = await API.get('/api/status');
      if (res.data.success && res.data.data) {
        setDocsLink(res.data.data.docs_link || '');
      }
    } catch (error) {
      // silent fail
    }
  };

  const fetchSkills = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/skills/');
      if (res.data.success) {
        const data = res.data.data || [];
        const skillList = Array.isArray(data) ? data : (data.items || []);
        setSkills(skillList);
      } else {
        showError(res.data.message || t('获取列表失败'));
      }
    } catch (error) {
      showError(t('获取列表失败'));
    } finally {
      setLoading(false);
    }
  };

  const fetchCategories = async () => {
    try {
      const res = await API.get('/api/skills/categories');
      if (res.data.success) {
        const data = res.data.data || [];
        const cats = Array.isArray(data)
          ? data.map((cat) => ({
              key: cat.name,
              label: cat.name,
              icon: <IconFile />,
            }))
          : Object.keys(data).map((key) => ({
              key,
              label: data[key] || key,
              icon: <IconFile />,
            }));
        setDynamicCategories([
          { key: 'all', label: t('全部'), icon: <IconGlobeStroke /> },
          ...cats,
        ]);
      }
    } catch (error) {
      // silent fail, use defaults
    }
  };

  const handleDownload = async (skill) => {
    if (downloadingId === skill.id) return;
    setDownloadingId(skill.id);

    try {
      // 直接使用下载接口，后端会返回文件流并增加下载计数
      window.open(`/api/skills/${skill.id}/download`, '_blank');
      showSuccess(t('下载成功'));
    } catch (error) {
      showError(t('下载失败'));
    } finally {
      setDownloadingId(null);
    }
  };

  const openDetail = (skill) => {
    setSelectedSkill(skill);
    setDetailVisible(true);
  };

  const closeDetail = () => {
    setDetailVisible(false);
    setSelectedSkill(null);
  };

  const handleSearch = () => {
    setActiveSearchQuery(searchQuery);
  };

  const filteredSkills = skills.filter((skill) => {
    const matchesSearch =
      !activeSearchQuery ||
      skill.name?.toLowerCase().includes(activeSearchQuery.toLowerCase()) ||
      skill.description?.toLowerCase().includes(activeSearchQuery.toLowerCase());
    const matchesCategory =
      selectedCategory === 'all' || skill.category === selectedCategory;
    return matchesSearch && matchesCategory && skill.status === 1;
  });

  // Dynamic category colors based on available categories
  const categoryColors = {};
  const categoryLabels = {};
  if (dynamicCategories) {
    const colors = ['blue', 'orange', 'purple', 'green', 'red', 'pink', 'teal', 'indigo', 'grey'];
    dynamicCategories.forEach((cat, idx) => {
      if (cat.key !== 'all') {
        categoryColors[cat.key] = colors[idx % colors.length];
        categoryLabels[cat.key] = cat.label;
      }
    });
  } else {
    categoryColors['other'] = 'grey';
    categoryLabels['other'] = t('其他');
  }

  return (
    <div style={{ minHeight: '100vh', background: 'var(--semi-color-bg-0)', marginTop: 64 }}>
      {/* 头部 Banner */}
      <div
        style={{
          background: 'linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%)',
          padding: isMobile ? '40px 20px' : '60px 40px',
          color: '#fff',
          textAlign: 'center',
        }}
      >
        <Title
          heading={isMobile ? 2 : 1}
          style={{ color: '#fff', marginBottom: 16 }}
        >
          {t('Skill 集市')}
        </Title>
        {docsLink && (
          <Paragraph
            style={{
              color: 'rgba(255,255,255,0.9)',
              fontSize: 16,
              maxWidth: 600,
              margin: '0 auto 24px',
              cursor: 'pointer',
              textDecoration: 'underline',
            }}
            onClick={() => window.open(docsLink, '_blank')}
          >
            {t('AI使用核心心法，请使用AI帮助你之前必读')}
          </Paragraph>
        )}

        {/* 搜索框 */}
        <div
          style={{
            maxWidth: 500,
            margin: '0 auto',
            display: 'flex',
            gap: 12,
          }}
        >
          <Input
            prefix={<IconSearch />}
            placeholder={t('搜索 Skill 名称或描述...')}
            value={searchQuery}
            onChange={(v) => setSearchQuery(v)}
            onEnterPress={handleSearch}
            size='large'
            style={{
              flex: 1,
              borderRadius: 8,
              background: 'rgba(255,255,255,0.95)',
            }}
          />
          <Button
            theme='solid'
            type='primary'
            icon={<IconSearch />}
            size='large'
            onClick={handleSearch}
            style={{
              borderRadius: 8,
            }}
          >
            {t('搜索')}
          </Button>
        </div>
      </div>

      <Content
        style={{
          maxWidth: 1200,
          margin: '0',
          padding: isMobile ? '16px' : '24px 40px 24px 0',
        }}
      >
        {/* 分类导航 */}
        <div
          style={{
            display: 'flex',
            gap: 8,
            flexWrap: 'wrap',
            marginBottom: 24,
            justifyContent: isMobile ? 'center' : 'flex-start',
          }}
        >
          {categories.map((cat) => (
            <Tag
              key={cat.key}
              size='large'
              color={selectedCategory === cat.key ? 'blue' : 'grey'}
              style={{
                cursor: 'pointer',
                padding: '8px 16px',
                fontSize: 14,
                borderRadius: 20,
              }}
              onClick={() => setSelectedCategory(cat.key)}
            >
              {cat.icon && (
                <span style={{ marginRight: 6 }}>{cat.icon}</span>
              )}
              {cat.label}
            </Tag>
          ))}
        </div>

        {/* 内容区域 */}
        {loading ? (
          <div style={{ textAlign: 'center', padding: 60 }}>
            <Spin size='large' />
          </div>
        ) : filteredSkills.length === 0 ? (
          <Empty
            title={t('暂无 Skill')}
            description={t('该分类下暂时没有可用的 Skill')}
            style={{ padding: 60 }}
          />
        ) : (
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: isMobile
                ? '1fr'
                : 'repeat(auto-fill, minmax(320px, 1fr))',
              gap: 20,
            }}
          >
            {filteredSkills.map((skill) => (
              <Card
                key={skill.id}
                shadows='hover'
                style={{
                  borderRadius: 12,
                  cursor: 'pointer',
                  transition: 'transform 0.2s',
                }}
                bodyStyle={{ padding: 20 }}
                onClick={() => openDetail(skill)}
              >
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'flex-start',
                    marginBottom: 12,
                  }}
                >
                  <div style={{ flex: 1 }}>
                    <Title heading={4} style={{ marginBottom: 8 }}>
                      {skill.name}
                    </Title>
                    <Tag
                      color={
                        categoryColors[skill.category] || 'grey'
                      }
                      size='small'
                    >
                      {categoryLabels[skill.category] || skill.category}
                    </Tag>
                  </div>
                </div>

                <Paragraph
                  ellipsis={{ rows: 2 }}
                  style={{ color: 'var(--semi-color-text-1)', minHeight: 44 }}
                >
                  {skill.description || t('暂无描述')}
                </Paragraph>

                <Divider margin={12} />

                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                  }}
                >
                  <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
                    <Text
                      size='small'
                      type='tertiary'
                      style={{ display: 'flex', alignItems: 'center', gap: 4 }}
                    >
                      <IconDownload size='small' />
                      {skill.downloads || 0}
                    </Text>
                    <Text
                      size='small'
                      type='tertiary'
                      style={{ display: 'flex', alignItems: 'center', gap: 4 }}
                    >
                      <IconCalendarClock size='small' />
                      {skill.updated_at
                        ? new Date(skill.updated_at).toLocaleDateString()
                        : ''}
                    </Text>
                  </div>

                  <Button
                    theme='solid'
                    type='primary'
                    icon={<IconDownload />}
                    loading={downloadingId === skill.id}
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDownload(skill);
                    }}
                  >
                    {t('下载')}
                  </Button>
                </div>
              </Card>
            ))}
          </div>
        )}
      </Content>

      {/* Skill 详情弹窗 */}
      <SideSheet
        title={selectedSkill?.name}
        visible={detailVisible}
        onCancel={closeDetail}
        width={isMobile ? '100%' : 560}
        placement='right'
      >
        {selectedSkill && (
          <div>

            <Title heading={3} style={{ marginBottom: 8 }}>
              {selectedSkill.name}
            </Title>

            <Tag
              color={categoryColors[selectedSkill.category] || 'grey'}
              style={{ marginBottom: 16 }}
            >
              {categoryLabels[selectedSkill.category] ||
                selectedSkill.category}
            </Tag>

            <Paragraph style={{ fontSize: 15, lineHeight: 1.8, marginBottom: 24 }}>
              {selectedSkill.description || t('暂无描述')}
            </Paragraph>

            {selectedSkill.detail && (
              <div style={{ marginBottom: 24 }}>
                <Title heading={5} style={{ marginBottom: 12 }}>
                  {t('详细介绍')}
                </Title>
                <Paragraph style={{ lineHeight: 1.8, whiteSpace: 'pre-wrap' }}>
                  {selectedSkill.detail}
                </Paragraph>
              </div>
            )}

            <Divider />

            <Descriptions
              data={[
                {
                  key: t('版本'),
                  value: selectedSkill.version || '1.0.0',
                },
                {
                  key: t('作者'),
                  value: selectedSkill.author || t('未知'),
                },
                {
                  key: t('下载次数'),
                  value: selectedSkill.downloads || 0
                },
                {
                  key: t('更新时间'),
                  value: selectedSkill.updated_at
                    ? new Date(selectedSkill.updated_at).toLocaleString()
                    : t('未知'),
                },
                {
                  key: t('文件大小'),
                  value: selectedSkill.file_size
                    ? `${(selectedSkill.file_size / 1024).toFixed(1)} KB`
                    : t('未知'),
                },
              ]}
              row
              size='small'
              style={{ marginBottom: 24 }}
            />

            <div style={{ display: 'flex', gap: 12 }}>
              <Button
                theme='solid'
                type='primary'
                icon={<IconDownload />}
                loading={downloadingId === selectedSkill.id}
                onClick={() => handleDownload(selectedSkill)}
                style={{ flex: 1 }}
              >
                {t('立即下载')}
              </Button>
            </div>

            <Banner
              type='info'
              fullMode={false}
              title={t('安装说明')}
              description={t(
                '下载后，在客户端的 Skill 管理页面中导入即可使用。'
              )}
              style={{ marginTop: 16 }}
            />
          </div>
        )}
      </SideSheet>
    </div>
  );
};

export default SkillMarketplace;