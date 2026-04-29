import React, { useEffect, useState, useRef, useCallback } from 'react';
import {
  Layout,
  Table,
  Button,
  Modal,
  Form,
  Input,
  TextArea,
  Select,
  Switch,
  Popconfirm,
  Tag,
  Toast,
  Upload,
  Spin,
  Space,
  Typography,
  Divider,
} from '@douyinfe/semi-ui';
import {
  IconPlus,
  IconDelete,
  IconEdit,
  IconUpload,
  IconSearch,
  IconRefresh,
  IconWrench,
  IconClose,
} from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import { useIsMobile } from '../../hooks/common/useIsMobile';

const { Content } = Layout;
const { Title, Text } = Typography;

const categoryColors = {
  产品经理类: 'blue',
  开发类: 'orange',
  测试类: 'purple',
  数据分析: 'green',
  通用类: 'grey',
};

const getCategoryColor = (cat) => categoryColors[cat] || 'grey';

// SkillAdmin - Admin page for skill management
const SkillAdmin = () => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const [skills, setSkills] = useState([]);
  const [loading, setLoading] = useState(true);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingSkill, setEditingSkill] = useState(null);
  const [uploadingFile, setUploadingFile] = useState(false);
  const [uploadedFile, setUploadedFile] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [categories, setCategories] = useState([]);
  const [categoryModalVisible, setCategoryModalVisible] = useState(false);
  const [editingCategory, setEditingCategory] = useState(null);
  const formRef = useRef(null);

  useEffect(() => {
    fetchSkills();
    fetchCategories();
  }, []);

  const fetchSkills = async () => {
    setLoading(true);
    try {
      const res = await API.get(`/api/skills/admin/?_t=${Date.now()}`);
      if (res.data.success) {
        // Backend returns data directly as array (not {items, total})
        const data = res.data.data || [];
        if (Array.isArray(data)) {
          setSkills(data);
        } else if (Array.isArray(data.items)) {
          setSkills(data.items);
        } else {
          setSkills([]);
        }
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
      const res = await API.get('/api/skills/admin/categories');
      if (res.data.success) {
        setCategories(res.data.data || []);
      }
    } catch (error) {
      // silent fail
    }
  };

  const handleDelete = async (id) => {
    try {
      const res = await API.delete(`/api/skills/admin/${id}`);
      if (res.data.success) {
        showSuccess(t('删除成功'));
        fetchSkills();
      } else {
        showError(res.data.message || t('删除失败'));
      }
    } catch (error) {
      showError(t('删除失败'));
    }
  };

  const handleSubmit = async (values) => {
    try {
      const payload = {
        name: values.name,
        description: values.description,
        detail: values.detail || '',
        category: values.category,
        version: values.version || '1.0.0',
        author: values.author || '',
        status: values.enabled ? 1 : 2,
        image_url: values.image_url || '',
        file_url: values.file_url || '',
        file_name: values.file_name || '',
      };

      let res;
      if (editingSkill) {
        // 编辑时需要包含 id
        payload.id = editingSkill.id;
        res = await API.put(`/api/skills/admin/${editingSkill.id}`, payload);
      } else {
        res = await API.post('/api/skills/admin/', payload);
      }

      if (res.data.success) {
        showSuccess(editingSkill ? t('更新成功') : t('创建成功'));
        setModalVisible(false);
        setEditingSkill(null);
        setUploadedFile(null);
        fetchSkills();
      } else {
        showError(res.data.message || t('操作失败'));
      }
    } catch (error) {
      showError(t('操作失败'));
    }
  };

  const openEdit = (skill) => {
    setEditingSkill(skill);
    setUploadedFile(
      skill.file_url
        ? { url: skill.file_url, name: skill.file_name || skill.file_url }
        : null
    );
    setModalVisible(true);
  };

  const openCreate = () => {
    setEditingSkill(null);
    setUploadedFile(null);
    setModalVisible(true);
  };

  const handleFileUpload = async (props) => {
    const file = props.fileInstance;
    if (!file) return;
    const fileList = [{ fileInstance: file }];
    const currentSkillId = editingSkill?.id;

    setUploadingFile(true);
    const formData = new FormData();
    formData.append('file', file);
    if (currentSkillId) {
      formData.append('skill_id', currentSkillId.toString());
    }

    try {
      const res = await API.post('/api/skills/admin/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      if (res.data.success) {
        showSuccess(t('上传成功'));
        const fileInfo = {
          url: res.data.data.url,
          name: res.data.data.filename || file.name,
        };
        setUploadedFile(fileInfo);
        formRef.current?.formApi?.setValues({
          ...formRef.current.formApi.getValues(),
          file_url: res.data.data.url,
          file_name: res.data.data.filename || file.name,
        });
        if (currentSkillId) {
          fetchSkills();
          const updatedSkill = {
            ...editingSkill,
            file_url: res.data.data.url,
            file_name: res.data.data.filename || file.name,
          };
          setEditingSkill(updatedSkill);
        }
      } else {
        showError(res.data.message || t('上传失败'));
      }
    } catch (error) {
      showError(t('上传失败'));
    } finally {
      setUploadingFile(false);
    }
  };


  // Category management
  const handleSaveCategory = async (values) => {
    try {
      let res;
      if (editingCategory) {
        res = await API.put(`/api/skills/admin/categories/${editingCategory.id}`, {
          id: editingCategory.id,
          name: values.name,
          sort: values.sort || 0,
        });
      } else {
        res = await API.post('/api/skills/admin/categories', {
          name: values.name,
          sort: values.sort || 0,
        });
      }
      if (res.data.success) {
        showSuccess(editingCategory ? t('更新成功') : t('创建成功'));
        setEditingCategory(null);
        fetchCategories();
      } else {
        showError(res.data.message || t('操作失败'));
      }
    } catch (error) {
      showError(t('操作失败'));
    }
  };

  const handleDeleteCategory = async (id) => {
    try {
      const res = await API.delete(`/api/skills/admin/categories/${id}`);
      if (res.data.success) {
        showSuccess(t('删除成功'));
        fetchCategories();
      } else {
        showError(res.data.message || t('删除失败'));
      }
    } catch (error) {
      showError(t('删除失败'));
    }
  };

  const filteredSkills = skills.filter(
    (s) =>
      !searchQuery ||
      s.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      s.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const columns = [
    {
      title: t('ID'),
      dataIndex: 'id',
      width: 60,
    },
    {
      title: t('名称'),
      dataIndex: 'name',
    },
    {
      title: t('分类'),
      dataIndex: 'category',
      render: (value) => (
        <Tag color={getCategoryColor(value)}>{value}</Tag>
      ),
    },
    {
      title: t('版本'),
      dataIndex: 'version',
      width: 90,
    },
    {
      title: t('下载次数'),
      dataIndex: 'downloads',
      width: 100,
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      render: (value) => (
        <Tag color={value === 1 ? 'green' : 'red'}>
          {value === 1 ? t('启用') : t('禁用')}
        </Tag>
      ),
      width: 80,
    },
    {
      title: t('操作'),
      dataIndex: 'operation',
      render: (_, record) => (
        <Space>
          <Button
            theme='light'
            type='primary'
            icon={<IconEdit />}
            onClick={() => openEdit(record)}
          >
            {t('编辑')}
          </Button>
          <Popconfirm
            title={t('确认删除')}
            content={t('删除后无法恢复，是否继续？')}
            onConfirm={() => handleDelete(record.id)}
            position='left'
          >
            <Button theme='light' type='danger' icon={<IconDelete />}>
              {t('删除')}
            </Button>
          </Popconfirm>
        </Space>
      ),
      width: 180,
    },
  ];

  const getInitialValues = useCallback(() => {
    if (editingSkill) {
      return {
        ...editingSkill,
        enabled: editingSkill.status === 1,
      };
    }
    return {
      enabled: true,
      category: categories.length > 0 ? categories[0].name : '',
      version: '1.0.0',
    };
  }, [editingSkill, categories]);

  return (
    <div style={{ padding: isMobile ? 16 : 24, marginTop: 64 }}>
      <Title heading={3} style={{ marginBottom: 24 }}>
        {t('Skill 管理')}
      </Title>

      <div
        style={{
          display: 'flex',
          gap: 12,
          marginBottom: 16,
          flexWrap: 'wrap',
          justifyContent: 'space-between',
        }}
      >
        <Space>
          <Button
            theme='solid'
            type='primary'
            icon={<IconPlus />}
            onClick={openCreate}
          >
            {t('新建 Skill')}
          </Button>
        </Space>

        <Space>
          <Input
            prefix={<IconSearch />}
            placeholder={t('搜索...')}
            value={searchQuery}
            onChange={(v) => setSearchQuery(v)}
            style={{ width: 200 }}
          />
          <Button icon={<IconRefresh />} onClick={fetchSkills}>
            {t('刷新')}
          </Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={filteredSkills}
        loading={loading}
        pagination={{ pageSize: 10 }}
        rowKey='id'
      />

      {/* 创建/编辑弹窗 */}
      <Modal
        title={editingSkill ? t('编辑 Skill') : t('新建 Skill')}
        visible={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          setEditingSkill(null);
          setUploadedFile(null);
        }}
        width={640}
        footer={null}
        maskClosable={false}
      >
        <Form
          ref={formRef}
          initValues={getInitialValues()}
          onSubmit={handleSubmit}
          labelPosition='top'
          style={{ paddingBottom: 15 }}
        >
          <Form.Input
            field='name'
            label={t('名称')}
            placeholder={t('Skill 名称')}
            rules={[{ required: true, message: t('请输入名称') }]}
          />

          {/* 分类 - Select + 管理按钮 */}
          <div style={{ marginBottom: 16 }}>
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: 8,
              }}
            >
              <Text type='secondary'>{t('分类')}</Text>
              <Button
                theme='borderless'
                type='tertiary'
                size='small'
                icon={<IconWrench />}
                onClick={() => setCategoryModalVisible(true)}
              >
                {t('管理分类')}
              </Button>
            </div>
            <Form.Select
              field='category'
              placeholder={t('选择分类')}
              rules={[{ required: true, message: t('请选择分类') }]}
              style={{ width: '100%' }}
              noLabel
            >
              {categories.map((cat) => (
                <Select.Option key={cat.id} value={cat.name}>
                  {cat.name}
                </Select.Option>
              ))}
            </Form.Select>
          </div>

          <Form.TextArea
            field='description'
            label={t('简介')}
            placeholder={t('简短描述')}
            rows={2}
            rules={[{ required: true, message: t('请输入简介') }]}
          />
          <Form.TextArea
            field='detail'
            label={t('详细介绍')}
            placeholder={t('详细描述（支持多行）')}
            rows={4}
          />
          <div
            style={{
              display: 'grid',
              gridTemplateColumns: '1fr 1fr',
              gap: 16,
            }}
          >
            <Form.Input
              field='version'
              label={t('版本')}
              placeholder='1.0.0'
            />
            <Form.Input
              field='author'
              label={t('作者')}
              placeholder={t('作者名称')}
            />
          </div>

          {/* Skill 文件上传 */}
          <div style={{ marginBottom: 16 }}>
            <Text type='secondary' style={{ display: 'block', marginBottom: 8 }}>
              {t('技能文件')}
            </Text>
            <Form.Input field='file_url' noLabel style={{ display: 'none' }} />
            <Space>
              <Upload
                action='#'
                showUploadList={false}
                customRequest={handleFileUpload}
                disabled={uploadingFile}
              >
                <Button
                  icon={<IconUpload />}
                  loading={uploadingFile}
                  theme='light'
                  type='tertiary'
                >
                  {t('上传文件')}
                </Button>
              </Upload>
              {(() => {
                const fileUrl = uploadedFile?.url || editingSkill?.file_url;
                const fileName =
                  uploadedFile?.name || editingSkill?.file_name;
                if (fileName || fileUrl) {
                  return (
                    <>
                      <Text
                        size='small'
                        type='tertiary'
                        style={{ maxWidth: 200 }}
                        ellipsis={{ showTooltip: true }}
                      >
                        {fileName || (typeof fileUrl === 'string' ? fileUrl.split('/').pop() : fileUrl)}
                      </Text>
                      <Button
                        theme='borderless'
                        type='danger'
                        size='small'
                        icon={<IconDelete />}
                        onClick={() => {
                          setUploadedFile(null);
                          formRef.current?.formApi?.setValues({
                            ...formRef.current.formApi.getValues(),
                            file_url: '',
                            file_name: '',
                          });
                        }}
                      >
                        {t('删除')}
                      </Button>
                    </>
                  );
                }
                return null;
              })()}
            </Space>
          </div>

          <Form.Switch
            field='enabled'
            label={t('启用')}
            size='large'
            style={{ width: 'auto', minWidth: 60 }}
          />
          <div
            style={{
              display: 'flex',
              justifyContent: 'flex-end',
              gap: 12,
              marginTop: 24,
              paddingBottom: 15,
            }}
          >
            <Button
              theme='light'
              type='tertiary'
              onClick={() => {
                setModalVisible(false);
                setEditingSkill(null);
                setUploadedFile(null);
              }}
            >
              {t('取消')}
            </Button>
            <Button theme='solid' type='primary' htmlType='submit'>
              {editingSkill ? t('保存') : t('创建')}
            </Button>
          </div>
        </Form>
      </Modal>

      {/* 分类管理弹窗 */}
      <Modal
        title={t('管理分类')}
        visible={categoryModalVisible}
        onCancel={() => {
          setCategoryModalVisible(false);
          setEditingCategory(null);
        }}
        footer={null}
        width={480}
      >
        <div style={{ paddingBottom: 15 }}>
          {/* 新建/编辑分类表单 */}
          <Form
            key={editingCategory ? `cat-${editingCategory.id}` : 'cat-new'}
            initValues={
              editingCategory || {
                name: '',
                sort: 0,
              }
            }
            onSubmit={handleSaveCategory}
            labelPosition='left'
            labelWidth={60}
          >
            <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
              <Form.Input
                field='name'
                label={t('名称')}
                placeholder={t('分类名称')}
                rules={[{ required: true, message: t('请输入分类名称') }]}
                style={{ flex: 1 }}
              />
              <Form.Input
                field='sort'
                label={t('排序')}
                placeholder='0'
                type='number'
                style={{ width: 100 }}
              />
              <Button theme='solid' type='primary' htmlType='submit'>
                {editingCategory ? t('更新') : t('添加')}
              </Button>
              {editingCategory && (
                <Button
                  theme='light'
                  type='tertiary'
                  icon={<IconClose />}
                  onClick={() => setEditingCategory(null)}
                />
              )}
            </div>
          </Form>

          <Divider margin={16} />

          {/* 分类列表 */}
          <div style={{ maxHeight: 300, overflowY: 'auto' }}>
            {categories.length === 0 ? (
              <Text type='tertiary' style={{ textAlign: 'center', display: 'block' }}>
                {t('暂无分类')}
              </Text>
            ) : (
              <Space vertical style={{ width: '100%' }}>
                {categories.map((cat) => (
                  <div
                    key={cat.id}
                    style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'center',
                      padding: '8px 12px',
                      borderRadius: 8,
                      background: 'var(--semi-color-fill-0)',
                    }}
                  >
                    <Space>
                      <Text strong>{cat.name}</Text>
                      <Text type='tertiary' size='small'>
                        {t('排序')}: {cat.sort}
                      </Text>
                    </Space>
                    <Space>
                      <Button
                        theme='borderless'
                        type='primary'
                        size='small'
                        icon={<IconEdit />}
                        onClick={() => setEditingCategory(cat)}
                      />
                      <Popconfirm
                        title={t('确认删除')}
                        content={t('删除后，使用该分类的 Skill 将变为无分类，是否继续？')}
                        onConfirm={() => handleDeleteCategory(cat.id)}
                      >
                        <Button
                          theme='borderless'
                          type='danger'
                          size='small'
                          icon={<IconDelete />}
                        />
                      </Popconfirm>
                    </Space>
                  </div>
                ))}
              </Space>
            )}
          </div>

          <div
            style={{
              display: 'flex',
              justifyContent: 'flex-end',
              marginTop: 16,
              paddingBottom: 15,
            }}
          >
            <Button
              theme='light'
              type='tertiary'
              onClick={() => {
                setCategoryModalVisible(false);
                setEditingCategory(null);
              }}
            >
              {t('关闭')}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default SkillAdmin;