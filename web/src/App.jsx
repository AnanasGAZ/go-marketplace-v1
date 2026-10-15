import { useEffect, useState } from 'react';

const ACCESS_TOKEN_KEY = 'access_token';
const TOKEN_TYPE_KEY = 'token_type';

async function postJSON(path, payload) {
  const response = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });

  const result = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(result.message || `Запрос завершился с ошибкой (${response.status})`);
  }

  return result;
}

function getAccessToken() {
  try {
    return sessionStorage.getItem(ACCESS_TOKEN_KEY);
  } catch {
    return null;
  }
}

function AuthForm() {
  const [login, setLogin] = useState('');
  const [password, setPassword] = useState('');
  const [pending, setPending] = useState(false);
  const [notice, setNotice] = useState(null);

  async function handleSubmit(event) {
    event.preventDefault();
    if (pending) return;

    const action = event.nativeEvent.submitter?.value || 'login';
    const credentials = { login, password };
    setPending(true);
    setNotice(null);

    try {
      if (action === 'register') {
        await postJSON('/api/v1/auth/register', {
          name: login,
          ...credentials,
        });
      }

      const result = await postJSON('/api/v1/auth/login', credentials);
      if (!result.access_token) {
        throw new Error('Сервер не вернул access token');
      }

      try {
        sessionStorage.setItem(ACCESS_TOKEN_KEY, result.access_token);
        sessionStorage.setItem(TOKEN_TYPE_KEY, result.token_type || 'Bearer');
      } catch {
        throw new Error('Браузер не разрешил сохранить сессию в этой вкладке');
      }

      window.location.assign(action === 'register' ? '/cabinet?showJwt=1' : '/cabinet');
    } catch (error) {
      setNotice({
        type: 'error',
        text: error instanceof Error ? error.message : 'Не удалось выполнить запрос',
      });
    } finally {
      setPending(false);
    }
  }

  return (
    <main className="screen">
      <section className="card" aria-labelledby="page-title">
        <div className="brand-mark" aria-hidden="true">M</div>
        <p className="eyebrow">Marketplace · тестовый интерфейс</p>
        <h1 id="page-title">Вход и регистрация</h1>
        <p className="intro">Введите логин и пароль, чтобы продолжить.</p>

        <form className="auth-form" onSubmit={handleSubmit}>
          <label className="field" htmlFor="login">
            <span>Логин</span>
            <input
              id="login"
              name="login"
              type="text"
              autoComplete="username"
              value={login}
              onChange={(event) => setLogin(event.target.value)}
              maxLength={200}
              required
            />
          </label>

          <label className="field" htmlFor="password">
            <span>Пароль</span>
            <input
              id="password"
              name="password"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              required
            />
          </label>

          <p className="hint">При регистрации логин также используется как имя пользователя.</p>

          {notice && (
            <p className={`notice ${notice.type}`} role="status" aria-live="polite">
              {notice.text}
            </p>
          )}

          <div className="actions">
            <button type="submit" value="login" disabled={pending}>
              {pending ? 'Подождите…' : 'Войти'}
            </button>
            <button className="secondary" type="submit" value="register" disabled={pending}>
              Регистрация
            </button>
          </div>
        </form>
      </section>
      <p className="footnote">После входа откроется заглушка личного кабинета.</p>
    </main>
  );
}

function Cabinet() {
  function logout() {
    try {
      sessionStorage.removeItem(ACCESS_TOKEN_KEY);
      sessionStorage.removeItem(TOKEN_TYPE_KEY);
    } catch {
      // Возвращаем пользователя на форму, даже если хранилище недоступно.
    }
    window.location.replace('/');
  }

  return (
    <main className="screen">
      <section className="card cabinet" aria-labelledby="cabinet-title">
        <p className="eyebrow">Marketplace</p>
        <h1 id="cabinet-title">Личный кабинет</h1>
        <p className="intro">Вы вошли в тестовый интерфейс.</p>
        <p className="placeholder">Тут будет личный кабинет.</p>
        {new URLSearchParams(window.location.search).get('showJwt') === '1' && (
          <section className="token-panel" aria-label="JWT после регистрации">
            <p className="token-heading">JWT <span>временно для отладки</span></p>
            <pre className="token-value">{getAccessToken()}</pre>
          </section>
        )}
        <button className="logout" type="button" onClick={logout}>Logout</button>
      </section>
    </main>
  );
}

export default function App() {
  const isCabinet = window.location.pathname === '/cabinet' || window.location.pathname === '/cabinet/';
  const hasToken = Boolean(getAccessToken());

  useEffect(() => {
    if (isCabinet && !hasToken) {
      window.location.replace('/');
    }
  }, [isCabinet, hasToken]);

  if (isCabinet) {
    return hasToken ? <Cabinet /> : null;
  }

  return <AuthForm />;
}
