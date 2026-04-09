import React, { useEffect, useMemo, useState } from "react";

const API_BASE = import.meta.env.VITE_API_BASE || "/api";

const initialRegister = {
  email: "",
  password: "",
  gender: "male",
  age: 25,
};

const initialLogin = {
  email: "",
  password: "",
};

const initialCreateProduct = {
  name: "",
  description: "",
  price: 0,
  stock: 0,
};

const initialUpdateProduct = {
  id: "",
  name: "",
  description: "",
  price: "",
  stock: "",
};

const initialUpdateUser = {
  id: "",
  email: "",
  age: "",
  gender: "",
  role: "",
};

function parseError(data, fallback) {
  if (data && typeof data === "object" && typeof data.error === "string") {
    return data.error;
  }

  return fallback;
}

function formatDate(value) {
  if (!value) return "-";
  return new Date(value).toLocaleString();
}

function readToken(key) {
  try {
    return localStorage.getItem(key) || "";
  } catch (_err) {
    return "";
  }
}

function writeToken(key, value) {
  try {
    localStorage.setItem(key, value);
  } catch (_err) {
    // Ignore localStorage failures in private mode.
  }
}

function clearToken(key) {
  try {
    localStorage.removeItem(key);
  } catch (_err) {
    // Ignore localStorage failures in private mode.
  }
}

function buildUpdatePayload(source, numericKeys = []) {
  const payload = {};

  Object.keys(source).forEach((key) => {
    if (key === "id") return;

    const value = source[key];
    if (value === "" || value === null || value === undefined) return;

    if (numericKeys.includes(key)) {
      payload[key] = Number(value);
      return;
    }

    payload[key] = value;
  });

  return payload;
}

function UserBadge({ role }) {
  return <span className={`badge role-${role || "guest"}`}>{role || "guest"}</span>;
}

function ProductCard({ product }) {
  return (
    <article className="item-card">
      <div className="item-head">
        <h4>{product.name}</h4>
        <span>#{product.id}</span>
      </div>
      <p>{product.description || "Без описания"}</p>
      <div className="meta-row">
        <span>Цена: {product.price}</span>
        <span>Остаток: {product.stock}</span>
        <span>Продавец: {product.owner_id}</span>
      </div>
    </article>
  );
}

function UserCard({ user }) {
  return (
    <article className="item-card">
      <div className="item-head">
        <h4>{user.email}</h4>
        <UserBadge role={user.role} />
      </div>
      <div className="meta-row">
        <span>ID: {user.id}</span>
        <span>Возраст: {user.age}</span>
        <span>Пол: {user.gender || "-"}</span>
        <span>Заблокирован: {user.is_blocked ? "Да" : "Нет"}</span>
      </div>
      <div className="meta-row">
        <span>Создан: {formatDate(user.created_at)}</span>
      </div>
    </article>
  );
}

export default function App() {
  const [screen, setScreen] = useState("home");
  const [status, setStatus] = useState("Готово к работе");
  const [busy, setBusy] = useState(false);

  const [registerForm, setRegisterForm] = useState(initialRegister);
  const [loginForm, setLoginForm] = useState(initialLogin);
  const [createProductForm, setCreateProductForm] = useState(initialCreateProduct);
  const [updateProductForm, setUpdateProductForm] = useState(initialUpdateProduct);
  const [updateUserForm, setUpdateUserForm] = useState(initialUpdateUser);

  const [productGetID, setProductGetID] = useState("");
  const [productDeleteID, setProductDeleteID] = useState("");
  const [userGetID, setUserGetID] = useState("");
  const [userBlockID, setUserBlockID] = useState("");

  const [accessToken, setAccessToken] = useState(() => readToken("access_token"));
  const [refreshToken, setRefreshToken] = useState(() => readToken("refresh_token"));
  const [currentUser, setCurrentUser] = useState(null);

  const [products, setProducts] = useState([]);
  const [selectedProduct, setSelectedProduct] = useState(null);
  const [users, setUsers] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);

  const isAuthenticated = Boolean(accessToken);
  const role = currentUser?.role || "guest";
  const isAdmin = role === "admin";
  const isSeller = role === "seller";
  const canCreateOrUpdateProduct = isSeller;

  const navItems = useMemo(
    () => [
      { key: "home", label: "Главная", enabled: true },
      { key: "auth", label: "Аккаунт", enabled: true },
      { key: "products", label: "Продукты", enabled: isAuthenticated },
      { key: "admin", label: "Админ", enabled: isAdmin },
    ],
    [isAuthenticated, isAdmin]
  );

  function authHeaders() {
    return accessToken ? { Authorization: `Bearer ${accessToken}` } : {};
  }

  async function callApi(path, options = {}) {
    const response = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...(options.headers || {}),
      },
    });

    let payload = null;
    try {
      payload = await response.json();
    } catch (_err) {
      payload = null;
    }

    if (!response.ok) {
      throw new Error(parseError(payload, `Ошибка запроса (${response.status})`));
    }

    return payload;
  }

  async function runAction(label, task) {
    setBusy(true);
    setStatus(`${label}...`);

    try {
      await task();
    } catch (err) {
      setStatus(`Ошибка: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  function applyAuth(data) {
    const nextAccess = data.access_token || "";
    const nextRefresh = data.refresh_token || "";

    setAccessToken(nextAccess);
    setRefreshToken(nextRefresh);
    setCurrentUser(data.user || null);

    writeToken("access_token", nextAccess);
    writeToken("refresh_token", nextRefresh);
  }

  function logout() {
    setAccessToken("");
    setRefreshToken("");
    setCurrentUser(null);
    setProducts([]);
    setUsers([]);
    setSelectedProduct(null);
    setSelectedUser(null);

    clearToken("access_token");
    clearToken("refresh_token");

    setScreen("home");
    setStatus("Вы вышли из аккаунта");
  }

  async function registerAndLogin() {
    return runAction("Регистрация", async () => {
      await callApi("/auth/register", {
        method: "POST",
        body: JSON.stringify({
          email: registerForm.email,
          password: registerForm.password,
          gender: registerForm.gender,
          age: Number(registerForm.age),
        }),
      });

      const authData = await callApi("/auth/login", {
        method: "POST",
        body: JSON.stringify({
          email: registerForm.email,
          password: registerForm.password,
        }),
      });

      applyAuth(authData);
      setScreen("products");
      setStatus("Вы успешно зарегистрировались и вошли в систему");
    });
  }

  async function login() {
    return runAction("Вход", async () => {
      const authData = await callApi("/auth/login", {
        method: "POST",
        body: JSON.stringify(loginForm),
      });

      applyAuth(authData);
      setScreen("products");
      setStatus("Добро пожаловать");
    });
  }

  async function refreshSession() {
    return runAction("Обновление сессии", async () => {
      const data = await callApi("/auth/refresh", {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      applyAuth(data);
      setStatus("Сессия обновлена");
    });
  }

  async function loadMe() {
    return runAction("Профиль", async () => {
      const data = await callApi("/auth/me", { headers: authHeaders() });
      setCurrentUser(data);
      setStatus("Профиль обновлен");
    });
  }

  async function loadProducts() {
    return runAction("Каталог", async () => {
      const data = await callApi("/products", { headers: authHeaders() });
      setProducts(Array.isArray(data) ? data : []);
      setStatus("Каталог загружен");
    });
  }

  async function getProductByID() {
    return runAction("Поиск продукта", async () => {
      const id = Number(productGetID);
      if (!id) {
        throw new Error("Введите корректный ID продукта");
      }

      const data = await callApi(`/products/${id}`, { headers: authHeaders() });
      setSelectedProduct(data);
      setStatus(`Продукт #${id} найден`);
    });
  }

  async function createProduct() {
    return runAction("Создание продукта", async () => {
      const data = await callApi("/products", {
        method: "POST",
        headers: authHeaders(),
        body: JSON.stringify({
          name: createProductForm.name,
          description: createProductForm.description,
          price: Number(createProductForm.price),
          stock: Number(createProductForm.stock),
        }),
      });

      setProducts((prev) => [data, ...prev]);
      setStatus("Продукт добавлен");
    });
  }

  async function updateProduct() {
    return runAction("Обновление продукта", async () => {
      const id = Number(updateProductForm.id);
      if (!id) {
        throw new Error("Введите ID продукта для изменения");
      }

      const payload = buildUpdatePayload(updateProductForm, ["price", "stock"]);
      if (Object.keys(payload).length === 0) {
        throw new Error("Заполните хотя бы одно поле для обновления");
      }

      const data = await callApi(`/products/${id}`, {
        method: "PUT",
        headers: authHeaders(),
        body: JSON.stringify(payload),
      });

      setSelectedProduct(data);
      setProducts((prev) => prev.map((item) => (item.id === data.id ? data : item)));
      setStatus(`Продукт #${id} обновлен`);
    });
  }

  async function deleteProduct() {
    return runAction("Удаление продукта", async () => {
      const id = Number(productDeleteID);
      if (!id) {
        throw new Error("Введите ID продукта для удаления");
      }

      await callApi(`/products/${id}`, {
        method: "DELETE",
        headers: authHeaders(),
      });

      setProducts((prev) => prev.filter((item) => item.id !== id));
      setSelectedProduct(null);
      setStatus(`Продукт #${id} удален`);
    });
  }

  async function loadUsers() {
    return runAction("Список пользователей", async () => {
      const data = await callApi("/users", { headers: authHeaders() });
      setUsers(Array.isArray(data) ? data : []);
      setStatus("Список пользователей обновлен");
    });
  }

  async function getUserByID() {
    return runAction("Поиск пользователя", async () => {
      const id = Number(userGetID);
      if (!id) {
        throw new Error("Введите корректный ID пользователя");
      }

      const data = await callApi(`/users/${id}`, { headers: authHeaders() });
      setSelectedUser(data);
      setStatus(`Пользователь #${id} найден`);
    });
  }

  async function updateUser() {
    return runAction("Обновление пользователя", async () => {
      const id = Number(updateUserForm.id);
      if (!id) {
        throw new Error("Введите ID пользователя для изменения");
      }

      const payload = buildUpdatePayload(updateUserForm, ["age"]);
      if (Object.keys(payload).length === 0) {
        throw new Error("Укажите хотя бы одно поле для обновления пользователя");
      }

      const data = await callApi(`/users/${id}`, {
        method: "PUT",
        headers: authHeaders(),
        body: JSON.stringify(payload),
      });

      setSelectedUser(data);
      setUsers((prev) => prev.map((item) => (item.id === data.id ? data : item)));
      setStatus(`Пользователь #${id} обновлен`);
    });
  }

  async function blockUser() {
    return runAction("Блокировка", async () => {
      const id = Number(userBlockID);
      if (!id) {
        throw new Error("Введите ID пользователя для блокировки");
      }

      await callApi(`/users/${id}`, {
        method: "DELETE",
        headers: authHeaders(),
      });

      setStatus(`Пользователь #${id} заблокирован`);
    });
  }

  useEffect(() => {
    if (!accessToken) return;

    loadMe();
  }, [accessToken]);

  function renderHome() {
    return (
      <section className="panel hero">
        <div>
          <p className="eyebrow">Интернет-магазин</p>
          <h1>Покупайте и управляйте товарами в одном месте</h1>
          <p className="lead">
            Добро пожаловать в магазин. Войдите в аккаунт, чтобы работать с каталогом и управлять товарами по своей роли.
          </p>
          <div className="actions">
            <button onClick={() => setScreen("products")} disabled={!isAuthenticated}>Перейти в каталог</button>
            <button className="ghost" onClick={() => setScreen("auth")}>{isAuthenticated ? "Сменить аккаунт" : "Войти или зарегистрироваться"}</button>
          </div>
        </div>

        <div className="hero-info">
          <h3>Ваш профиль</h3>
          <p><strong>Пользователь:</strong> {currentUser?.email || "гость"}</p>
          <p><strong>Ваша роль:</strong> <UserBadge role={role} /></p>
          <p><strong>Права:</strong> {role === "seller" ? "создание и редактирование товаров" : role === "admin" ? "админ-панель и удаление товаров" : role === "user" ? "просмотр каталога" : "ограниченный доступ"}</p>
          <p><strong>Статус:</strong> {status}</p>
          {isAuthenticated ? <button className="ghost" disabled={busy} onClick={loadMe}>Обновить роль и профиль</button> : null}
        </div>
      </section>
    );
  }

  function renderAuth() {
    return (
      <section className="panel">
        <h2>Аккаунт</h2>
        <p className="muted">После регистрации вход выполняется автоматически.</p>

        <div className="two-columns">
          <div className="card">
            <h3>Регистрация</h3>
            <label>
              Email
              <input value={registerForm.email} onChange={(e) => setRegisterForm({ ...registerForm, email: e.target.value })} />
            </label>
            <label>
              Password
              <input type="password" value={registerForm.password} onChange={(e) => setRegisterForm({ ...registerForm, password: e.target.value })} />
            </label>
            <label>
              Gender
              <input value={registerForm.gender} onChange={(e) => setRegisterForm({ ...registerForm, gender: e.target.value })} />
            </label>
            <label>
              Age
              <input type="number" value={registerForm.age} onChange={(e) => setRegisterForm({ ...registerForm, age: e.target.value })} />
            </label>
            <button disabled={busy} onClick={registerAndLogin}>Создать аккаунт</button>
          </div>

          <div className="card">
            <h3>Вход</h3>
            <label>
              Email
              <input value={loginForm.email} onChange={(e) => setLoginForm({ ...loginForm, email: e.target.value })} />
            </label>
            <label>
              Password
              <input type="password" value={loginForm.password} onChange={(e) => setLoginForm({ ...loginForm, password: e.target.value })} />
            </label>
            <div className="actions compact">
              <button disabled={busy} onClick={login}>Войти</button>
              <button className="ghost" disabled={busy || !refreshToken} onClick={refreshSession}>Обновить сессию</button>
            </div>
            {isAuthenticated ? (
              <div className="actions compact">
                <button className="ghost" disabled={busy} onClick={loadMe}>Обновить профиль</button>
                <button className="ghost" disabled={busy} onClick={logout}>Выйти</button>
              </div>
            ) : null}
          </div>
        </div>
      </section>
    );
  }

  function renderProducts() {
    if (!isAuthenticated) {
      return (
        <section className="panel">
          <h2>Каталог товаров</h2>
          <p className="muted">Для доступа к каталогу нужно войти в аккаунт.</p>
          <button onClick={() => setScreen("auth")}>Перейти ко входу</button>
        </section>
      );
    }

    return (
      <section className="panel">
        <div className="section-head">
          <div>
            <h2>Каталог товаров</h2>
            <p className="muted">Просматривайте товары и управляйте ими согласно своей роли.</p>
          </div>
          <button disabled={busy} onClick={loadProducts}>Обновить каталог</button>
        </div>

        <div className="selected-card top-gap">
          <p><strong>Текущая роль:</strong> <UserBadge role={role} /></p>
          <p>
            {isSeller
              ? "Вы можете создавать и редактировать товары."
              : isAdmin
                ? "Вы можете удалять товары и управлять пользователями."
                : "Для создания и редактирования товаров нужна роль seller."}
          </p>
        </div>

        <div className="two-columns">
          <div className="card">
            <h3>Поиск товара</h3>
            <label>
              ID товара
              <input value={productGetID} onChange={(e) => setProductGetID(e.target.value)} />
            </label>
            <div className="actions compact">
              <button disabled={busy} onClick={getProductByID}>Найти товар</button>
              <button className="ghost" disabled={busy || !isAdmin} onClick={deleteProduct}>Удалить (admin)</button>
            </div>
            <label className="top-gap">
              ID для удаления
              <input value={productDeleteID} onChange={(e) => setProductDeleteID(e.target.value)} />
            </label>

            {selectedProduct ? (
              <div className="selected-card top-gap">
                <h4>Найденный товар</h4>
                <p><strong>{selectedProduct.name}</strong></p>
                <p>{selectedProduct.description}</p>
                <p>Цена: {selectedProduct.price}, Остаток: {selectedProduct.stock}</p>
              </div>
            ) : null}
          </div>

          <div className="card">
            <h3>Создание товара</h3>
            <label>
              Название
              <input value={createProductForm.name} onChange={(e) => setCreateProductForm({ ...createProductForm, name: e.target.value })} />
            </label>
            <label>
              Описание
              <input value={createProductForm.description} onChange={(e) => setCreateProductForm({ ...createProductForm, description: e.target.value })} />
            </label>
            <label>
              Цена
              <input type="number" step="0.01" value={createProductForm.price} onChange={(e) => setCreateProductForm({ ...createProductForm, price: e.target.value })} />
            </label>
            <label>
              Количество
              <input type="number" value={createProductForm.stock} onChange={(e) => setCreateProductForm({ ...createProductForm, stock: e.target.value })} />
            </label>
            <button disabled={busy || !canCreateOrUpdateProduct} onClick={createProduct}>Добавить товар (seller)</button>
          </div>
        </div>

        <div className="card top-gap">
          <h3>Редактирование товара</h3>
          <div className="grid">
            <label>
              ID товара
              <input value={updateProductForm.id} onChange={(e) => setUpdateProductForm({ ...updateProductForm, id: e.target.value })} />
            </label>
            <label>
              Название
              <input value={updateProductForm.name} onChange={(e) => setUpdateProductForm({ ...updateProductForm, name: e.target.value })} />
            </label>
            <label>
              Описание
              <input value={updateProductForm.description} onChange={(e) => setUpdateProductForm({ ...updateProductForm, description: e.target.value })} />
            </label>
            <label>
              Цена
              <input value={updateProductForm.price} onChange={(e) => setUpdateProductForm({ ...updateProductForm, price: e.target.value })} />
            </label>
            <label>
              Количество
              <input value={updateProductForm.stock} onChange={(e) => setUpdateProductForm({ ...updateProductForm, stock: e.target.value })} />
            </label>
          </div>
          <button disabled={busy || !canCreateOrUpdateProduct} onClick={updateProduct}>Сохранить изменения (seller)</button>
        </div>

        <div className="cards-grid top-gap">
          {products.length === 0 ? <div className="empty">Пока нет загруженных товаров.</div> : products.map((product) => <ProductCard key={product.id} product={product} />)}
        </div>
      </section>
    );
  }

  function renderAdmin() {
    if (!isAuthenticated) {
      return (
        <section className="panel">
          <h2>Админ-панель</h2>
          <p className="muted">Сначала войдите в аккаунт.</p>
        </section>
      );
    }

    if (!isAdmin) {
      return (
        <section className="panel">
          <h2>Админ-панель</h2>
          <p className="muted">Этот раздел доступен только для роли admin.</p>
        </section>
      );
    }

    return (
      <section className="panel">
        <div className="section-head">
          <div>
            <h2>Админ-панель</h2>
            <p className="muted">Управление пользователями: просмотр, изменение и блокировка.</p>
          </div>
          <button disabled={busy} onClick={loadUsers}>Обновить список</button>
        </div>

        <div className="two-columns">
          <div className="card">
            <h3>Поиск и блокировка</h3>
            <label>
              ID пользователя
              <input value={userGetID} onChange={(e) => setUserGetID(e.target.value)} />
            </label>
            <div className="actions compact">
              <button disabled={busy} onClick={getUserByID}>Найти пользователя</button>
            </div>
            <label className="top-gap">
              ID для блокировки
              <input value={userBlockID} onChange={(e) => setUserBlockID(e.target.value)} />
            </label>
            <button className="ghost" disabled={busy} onClick={blockUser}>Заблокировать</button>

            {selectedUser ? (
              <div className="selected-card top-gap">
                <h4>Профиль пользователя</h4>
                <p><strong>{selectedUser.email}</strong></p>
                <p>Роль: {selectedUser.role}</p>
                <p>Возраст: {selectedUser.age}, Пол: {selectedUser.gender || "-"}</p>
              </div>
            ) : null}
          </div>

          <div className="card">
            <h3>Редактирование пользователя</h3>
            <label>
              ID
              <input value={updateUserForm.id} onChange={(e) => setUpdateUserForm({ ...updateUserForm, id: e.target.value })} />
            </label>
            <label>
              Email
              <input value={updateUserForm.email} onChange={(e) => setUpdateUserForm({ ...updateUserForm, email: e.target.value })} />
            </label>
            <label>
              Возраст
              <input value={updateUserForm.age} onChange={(e) => setUpdateUserForm({ ...updateUserForm, age: e.target.value })} />
            </label>
            <label>
              Пол
              <input value={updateUserForm.gender} onChange={(e) => setUpdateUserForm({ ...updateUserForm, gender: e.target.value })} />
            </label>
            <label>
              Роль (user/seller/admin)
              <input value={updateUserForm.role} onChange={(e) => setUpdateUserForm({ ...updateUserForm, role: e.target.value })} />
            </label>
            <button disabled={busy} onClick={updateUser}>Сохранить пользователя</button>
          </div>
        </div>

        <div className="cards-grid top-gap">
          {users.length === 0 ? <div className="empty">Список пользователей пуст.</div> : users.map((user) => <UserCard key={user.id} user={user} />)}
        </div>
      </section>
    );
  }

  function renderScreen() {
    if (screen === "auth") return renderAuth();
    if (screen === "products") return renderProducts();
    if (screen === "admin") return renderAdmin();
    return renderHome();
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <div className="brand-block">
          <span className="brand">somewebproject</span>
          <span>Роль:</span>
          <UserBadge role={role} />
        </div>

        <nav className="nav">
          {navItems.map((item) => (
            <button
              key={item.key}
              className={`nav-btn ${screen === item.key ? "active" : ""}`}
              disabled={!item.enabled || busy}
              onClick={() => setScreen(item.key)}
            >
              {item.label}
            </button>
          ))}
        </nav>

        <div className="top-actions">
          <button className="ghost" disabled={!isAuthenticated || busy} onClick={loadMe}>Обновить профиль</button>
          <button className="ghost" disabled={!isAuthenticated || busy || !refreshToken} onClick={refreshSession}>Refresh</button>
          {isAuthenticated ? <button className="danger" disabled={busy} onClick={logout}>Выйти</button> : null}
        </div>
      </header>

      <div className={`notice ${status.startsWith("Ошибка") ? "error" : "ok"}`}>{status}</div>

      {renderScreen()}
    </main>
  );
}
