import React from 'react';
import { Layout as AntLayout, Button, Space, Dropdown, Avatar, Tooltip, Grid } from 'antd';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import {
  BookOutlined,
  DashboardOutlined,
  UserOutlined,
  LogoutOutlined,
  DownOutlined,
  BulbOutlined,
  BulbFilled,
  GlobalOutlined,
  SafetyOutlined,
} from '@ant-design/icons';
import { useAuth } from '@/contexts/AuthContext';
import { useThemeMode } from '@/contexts/ThemeContext';
import { useTranslation } from 'react-i18next';

const { Header, Content, Footer } = AntLayout;
const { useBreakpoint } = Grid;

interface LayoutProps {
  children: React.ReactNode;
}

// The application shell: a top navigation bar, the routed content area and a
// footer. The bar carries primary navigation, the colour-mode toggle, the
// language switch and the account menu. It collapses gracefully on tablet
// widths by hiding the inline nav links behind the responsive breakpoint.
export const Layout: React.FC<LayoutProps> = ({ children }) => {
  const { user, role, isAuthenticated, logout } = useAuth();
  const { mode, toggle } = useThemeMode();
  const navigate = useNavigate();
  const location = useLocation();
  const { t, i18n } = useTranslation();
  const screens = useBreakpoint();

  // `md` (>=768px) is our tablet threshold: above it we show inline nav links,
  // at or below it the account dropdown remains the primary affordance.
  const showInlineNav = screens.md;

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng);
    localStorage.setItem('language', lng);
  };

  const isActive = (path: string) =>
    location.pathname === path || location.pathname.startsWith(`${path}/`);

  const accountMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: t('nav.profile'),
      onClick: () => navigate('/profile'),
    },
    ...(role === 'root'
      ? [
          {
            key: 'admin',
            icon: <SafetyOutlined />,
            label: t('nav.admin'),
            onClick: () => navigate('/admin'),
          },
        ]
      : []),
    { type: 'divider' as const },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: t('nav.logout'),
      onClick: handleLogout,
    },
  ];

  return (
    <AntLayout style={{ minHeight: '100vh' }}>
      <Header
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 clamp(16px, 4vw, 50px)',
          gap: 16,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '40px', minWidth: 0 }}>
          <Link
            to="/dashboard"
            style={{ color: 'white', fontSize: '20px', fontWeight: 'bold', whiteSpace: 'nowrap' }}
          >
            InfoMark
          </Link>
          {isAuthenticated && showInlineNav && (
            <Space size="large">
              <Link
                to="/dashboard"
                aria-current={isActive('/dashboard') ? 'page' : undefined}
                style={{ color: isActive('/dashboard') ? '#c0b3ff' : 'white', fontSize: '16px' }}
              >
                <DashboardOutlined /> {t('nav.dashboard')}
              </Link>
              <Link
                to="/courses"
                aria-current={isActive('/courses') ? 'page' : undefined}
                style={{ color: isActive('/courses') ? '#c0b3ff' : 'white', fontSize: '16px' }}
              >
                <BookOutlined /> {t('nav.courses')}
              </Link>
            </Space>
          )}
        </div>

        <Space size="middle">
          {/* Colour-mode toggle: persisted via ThemeContext/localStorage. */}
          <Tooltip title={mode === 'dark' ? t('nav.lightMode') : t('nav.darkMode')}>
            <Button
              type="text"
              aria-label={mode === 'dark' ? t('nav.lightMode') : t('nav.darkMode')}
              icon={
                mode === 'dark' ? (
                  <BulbFilled style={{ color: 'white', fontSize: 18 }} />
                ) : (
                  <BulbOutlined style={{ color: 'white', fontSize: 18 }} />
                )
              }
              onClick={toggle}
            />
          </Tooltip>

          <Dropdown
            menu={{
              items: [
                { key: 'en', label: 'English', onClick: () => changeLanguage('en') },
                { key: 'de', label: 'Deutsch', onClick: () => changeLanguage('de') },
              ],
            }}
          >
            <Button
              type="text"
              aria-label={t('nav.language')}
              icon={<GlobalOutlined style={{ color: 'white', fontSize: 18 }} />}
            >
              <span style={{ color: 'white', textTransform: 'uppercase' }}>{i18n.language.slice(0, 2)}</span>
            </Button>
          </Dropdown>

          {isAuthenticated && user ? (
            <Dropdown menu={{ items: accountMenuItems }}>
              <Space style={{ cursor: 'pointer', color: 'white' }} tabIndex={0} role="button" aria-label={t('nav.profile')}>
                <Avatar src={user.avatar_url} icon={!user.avatar_url && <UserOutlined />} />
                {showInlineNav && (
                  <span>
                    {user.first_name} {user.last_name}
                  </span>
                )}
                <DownOutlined />
              </Space>
            </Dropdown>
          ) : (
            <Space>
              <Button type="default" onClick={() => navigate('/login')}>
                {t('auth.login')}
              </Button>
              <Button type="primary" onClick={() => navigate('/register')}>
                {t('auth.register')}
              </Button>
            </Space>
          )}
        </Space>
      </Header>

      <Content style={{ padding: 'clamp(16px, 3vw, 24px) clamp(12px, 4vw, 50px)' }}>{children}</Content>

      <Footer style={{ textAlign: 'center' }}>
        <div>{t('footer.copyright')}</div>
        <Space style={{ marginTop: 8 }} wrap>
          <Link to="/terms">{t('footer.terms')}</Link>
          <a href="https://github.com/infomark-org" target="_blank" rel="noopener noreferrer">
            {t('footer.github')}
          </a>
        </Space>
      </Footer>
    </AntLayout>
  );
};
