import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/authStore'
import LoginForm from '../components/LoginForm.vue'
import Dashboard from '../components/Dashboard.vue'

// Extend Vue Router types to include meta information
declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean // Does this route need login?
  }
}

// Create router instance
const router = createRouter({
  // Use browser history (normal URLs without #)
  history: createWebHistory(import.meta.env.BASE_URL),

  // Define all routes
  routes: [
    // ========== ROOT ROUTE ==========
    // When user goes to "/"
    // Redirect to login or dashboard based on login status
    {
      path: '/',
      redirect: () => {
        const authStore = useAuthStore()
        // If logged in, go to dashboard
        // If not logged in, go to login
        return authStore.isAuthenticated ? '/dashboard' : '/login'
      },
    },

    // ========== LOGIN ROUTE ==========
    // When user goes to "/login"
    {
      path: '/login',
      name: 'Login',
      component: LoginForm,
      meta: {
        requiresAuth: false, // Anyone can access this
      },
    },
    // ========== REGISTER ROUTE ==========
    // When user goes to "/register"
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/components/RegisterForm.vue'),
      meta: {
        requiresAuth: false, // Anyone can access this
      },
    },
    // ========== DASHBOARD ROUTE ==========
    // When user goes to "/dashboard"
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: Dashboard,
      meta: {
        requiresAuth: true, // Only logged-in users can access
      },
    },

    // ========== OAUTH CALLBACK ROUTE ==========
    // When Auth0 redirects back after login
    {
      path: '/oauth/callback',
      name: 'OAuthCallback',
      component: () => import('@/components/OAuthCallback.vue'),
      meta: {
        requiresAuth: false, // Callback can be accessed without login
      },
    },

    // ========== 404 ROUTE ==========
    // If user goes to a URL that doesn't exist
    {
      path: '/:pathMatch(.*)*',
      redirect: '/login',
    },
  ],
})

// ========== NAVIGATION GUARD ==========
// Run BEFORE each navigation
// Check if user has permission to access the route
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  // On first navigation, restore user from localStorage
  if (!authStore.user && !authStore.accessToken) {
    authStore.initializeAuth()
  }

  // Does this route require authentication?
  const requiresAuth = to.meta.requiresAuth ?? false

  // Is the user logged in?
  const isAuthenticated = authStore.isAuthenticated

  // Log for debugging
  console.log(`Navigation: ${from.path} → ${to.path}`, {
    requiresAuth,
    isAuthenticated,
    user: authStore.user?.email,
  })

  // ========== PROTECTED ROUTE ==========
  // Route requires auth AND user is not logged in
  if (requiresAuth && !isAuthenticated) {
    console.warn('Access denied: User not authenticated')
    // Redirect to login
    next('/login')
    return
  }

  // ========== REDIRECT LOGGED-IN USERS ==========
  // User is logged in AND trying to access login page
  if (to.path === '/login' && isAuthenticated) {
    console.log('User already authenticated, redirecting to dashboard')
    // Redirect to dashboard
    next('/dashboard')
    return
  }
  // Redirect logged-in users away from register page
  if (to.path === '/register' && isAuthenticated) {
    console.log('User already authenticated, redirecting to dashboard')
    next('/dashboard')
    return
  }
  // User is logged in AND trying to access OAuth callback
  if (to.path === '/oauth/callback' && isAuthenticated) {
    console.log('User already authenticated, redirecting to dashboard')
    next('/dashboard')
    return
  }

  // Allow navigation
  next()
})

// ========== AFTER NAVIGATION ==========
// Run AFTER each navigation
// Update page title and scroll to top
router.afterEach((to, from) => {
  // Set page title based on route name
  const titles: Record<string, string> = {
    Login: 'Login - Prost',
    Register: 'Register - Prost',
    Dashboard: 'Dashboard - Prost',
    OAuthCallback: 'OAuth Callback - Prost',
  }

  document.title = titles[to.name as string] || 'Prost'

  // Scroll to top of page
  window.scrollTo(0, 0)
})

// ========== ERROR HANDLING ==========
// If something goes wrong during navigation
router.onError((error) => {
  console.error('Router error:', error)
})

export default router
