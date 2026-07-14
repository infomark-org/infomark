import React, { useEffect, useState } from 'react';
import { Card, Result, Button, Spin } from 'antd';
import { useParams, useNavigate } from 'react-router-dom';
import agent from '@/api/agent';
import { useTranslation } from 'react-i18next';

const MailConfirmation: React.FC = () => {
  // The backend confirmation email links to `/#/confirmation/{email}/{token}`
  // (account.go + email/email.go), so both values arrive as path parameters.
  const { email, token } = useParams<{ email: string; token: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const [loading, setLoading] = useState(true);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (token && email) {
      confirmEmail();
    } else {
      setError(t('confirmation.invalidLink'));
      setLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token, email]);

  const confirmEmail = async () => {
    try {
      // `ConfirmEmailRequest.Bind` (auth_requests.go) requires both email and
      // confirmation_token.
      await agent.Auth.confirmEmail({
        email: decodeURIComponent(email!),
        confirmation_token: token!,
      });
      setSuccess(true);
    } catch (err: any) {
      setError(err.message || t('confirmation.confirmFailed'));
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '70vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '70vh' }}>
      <Card style={{ width: 500 }}>
        {success ? (
          <Result
            status="success"
            title={t('confirmation.successTitle')}
            subTitle={t('confirmation.successSubtitle')}
            extra={[
              <Button type="primary" key="login" onClick={() => navigate('/login')}>
                {t('confirmation.goToLogin')}
              </Button>,
            ]}
          />
        ) : (
          <Result
            status="error"
            title={t('confirmation.failedTitle')}
            subTitle={error || t('confirmation.confirmFailed')}
            extra={[
              <Button type="primary" key="home" onClick={() => navigate('/')}>
                {t('confirmation.goHome')}
              </Button>,
            ]}
          />
        )}
      </Card>
    </div>
  );
};

export default MailConfirmation;
