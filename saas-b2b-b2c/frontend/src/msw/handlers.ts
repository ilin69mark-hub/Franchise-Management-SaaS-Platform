import { http, HttpResponse, delay } from 'msw';
import { setupServer } from 'msw/node';

interface User {
  id: string;
  email: string;
  role: string;
  first_name?: string;
  last_name?: string;
}

interface AuthResponse {
  user: User;
  token: string;
  refresh_token: string;
}

const mockUsers: User[] = [
  { id: 'user-1', email: 'dealer@example.com', role: 'dealer', first_name: 'John', last_name: 'Doe' },
  { id: 'user-2', email: 'franchiser@example.com', role: 'franchiser', first_name: 'Jane', last_name: 'Smith' },
];

const mockDealers = [
  { id: 'dealer-1', name: 'Salon Moscow', email: 'moscow@salon.com', phone: '+79991234567', status: 'active' },
  { id: 'dealer-2', name: 'Salon SPB', email: 'spb@salon.com', phone: '+79999876543', status: 'active' },
];

export const handlers = [
  http.post('/api/v1/auth/login', async ({ request }) => {
    await delay(200);
    const body = await request.json() as { email: string; password: string };
    
    const user = mockUsers.find(u => u.email === body.email);
    if (!user) {
      return HttpResponse.json(
        { error: 'invalid credentials' },
        { status: 401 }
      );
    }

    return HttpResponse.json({
      user,
      token: 'mock_access_token_' + user.id,
      refresh_token: 'mock_refresh_token_' + user.id,
    });
  }),

  http.post('/api/v1/auth/register', async ({ request }) => {
    await delay(200);
    const body = await request.json() as { email: string; password: string; role: string };
    
    const newUser: User = {
      id: 'user-' + Date.now(),
      email: body.email,
      role: body.role || 'dealer',
    };

    return HttpResponse.json({
      user: newUser,
      token: 'mock_access_token_' + newUser.id,
      refresh_token: 'mock_refresh_token_' + newUser.id,
    }, { status: 201 });
  }),

  http.get('/api/v1/auth/me', ({ request }) => {
    const authHeader = request.headers.get('Authorization');
    if (!authHeader?.startsWith('Bearer ')) {
      return HttpResponse.json({ error: 'unauthorized' }, { status: 401 });
    }

    return HttpResponse.json({ user: mockUsers[0] });
  }),

  http.get('/api/v1/dealers', ({ request }) => {
    const authHeader = request.headers.get('Authorization');
    if (!authHeader?.startsWith('Bearer ')) {
      return HttpResponse.json({ error: 'unauthorized' }, { status: 401 });
    }

    return HttpResponse.json({ dealers: mockDealers });
  }),

  http.get('/api/v1/dealers/:id', ({ params }) => {
    const dealer = mockDealers.find(d => d.id === params.id);
    if (!dealer) {
      return HttpResponse.json({ error: 'not found' }, { status: 404 });
    }
    return HttpResponse.json({ dealer });
  }),

  http.get('/api/v1/products', () => {
    return HttpResponse.json({
      products: [
        { id: 'prod-1', name: 'Product A', price: 1000 },
        { id: 'prod-2', name: 'Product B', price: 2000 },
      ],
    });
  }),

  http.post('/api/v1/orders', async ({ request }) => {
    await delay(200);
    const body = await request.json();
    
    return HttpResponse.json({
      id: 'order-' + Date.now(),
      status: 'pending',
      ...body,
    }, { status: 201 });
  }),

  http.get('/api/v1/error/500', () => {
    return HttpResponse.json(
      { error: 'Internal Server Error', message: 'Something went wrong' },
      { status: 500 }
    );
  }),

  http.get('/api/v1/error/403', () => {
    return HttpResponse.json(
      { error: 'Forbidden', message: 'Access denied' },
      { status: 403 }
    );
  }),
];

export const server = setupServer(...handlers);