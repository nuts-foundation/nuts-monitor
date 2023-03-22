import { createApp } from 'vue'
import { createRouter, createWebHashHistory } from 'vue-router'
import './style.css'
import App from './App.vue'
import NotFound from './NotFound.vue'
import Landing from './Landing.vue'

const routes = [
  { path: '/', component: Landing },
  { path: '/:pathMatch*', name: 'NotFound', component: NotFound }
]

const router = createRouter({
  // We are using the hash history for simplicity here.
  history: createWebHashHistory(),
  routes // short for `routes: routes`
})

router.beforeEach((to, from) => {
  // apply security
})

const app = createApp(App)

app.use(router)
app.mount('#app')
