import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import './style.css'
import Api from './plugins/api'
import App from './App.vue'
import AdminApp from './admin/AdminApp.vue'
import Diagnostics from './admin/Diagnostics.vue'
import Logout from './Logout.vue'
import NetworkTopology from './admin/NetworkTopology.vue'
import Transactions from './admin/Transactions.vue'
import NotFound from './NotFound.vue'

// Vuetify
import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

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
      },
      {
        path: 'network_topology',
        name: 'admin.network_topology',
        component: NetworkTopology
      },
      {
        path: 'transactions',
        name: 'admin.transactions',
        component: Transactions
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

const vuetify = createVuetify({
  components,
  directives,
})

app.use(router)
app.use(vuetify)
app.use(Api, { forbiddenRoute: { name: 'logout' } })
app.mount('#app')
