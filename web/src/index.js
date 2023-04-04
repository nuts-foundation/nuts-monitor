import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import './style.css'
import Api from './plugins/api'
import App from './App.vue'
import AdminApp from './admin/AdminApp.vue'
import Diagnostics from './admin/Diagnostics.vue'
import Logout from './Logout.vue'
import NotFound from './NotFound.vue'

const routes = [
  {
    name: 'logout',
    path: '/logout',
    component: Logout
  },
  {
    path: '/',
    components: {
      default: AdminApp
    },
    children: [
      {
        path: '',
        name: 'admin.home',
        redirect: '/diagnostics'
      },
      {
        path: 'diagnostics',
        name: 'admin.diagnostics',
        component: Diagnostics
      }
    ],
    //meta: { requiresAuth: true }
  },

  { path: '/:pathMatch*', name: 'NotFound', component: NotFound }
]

const router = createRouter({
  // We are using the hash history for simplicity here.
  history: createWebHashHistory(),
  routes // short for `routes: routes`
})

router.beforeEach(() => {
  // apply security
})

const app = createApp(App)

app.use(router)
app.use(Api, { forbiddenRoute: { name: 'logout' } })
app.mount('#app')
