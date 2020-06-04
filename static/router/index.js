import Vue from '../vendor/vue/vue.js'
import Router from '../vendor/vue/vue-router.js'
import Login from '../components/Login.vue'
import Home from '../components/Home.vue'
import '../assets/css/global.css'

Vue.use(Router)

const router = new Router({
  routes: [
    { path: '/', redirect: '/login' },
    { path: '/login', component: Login },
    { path: '/home', component: Home }
  ]
})

router.beforeEach(
  (to, from, next) => {
    // to 将要访问的路径
    // from 代表从哪个路径而来
    // next 是一个函数，表示放行
    // next() 放行  next('/login') 强制跳转
    if (to.path === '/login') return next()
    // 获取 token
    const tokenStr = window.sessionStorage.getItem('token')
    if (!tokenStr) return next('/login')
    next()
  }
)

export default router
