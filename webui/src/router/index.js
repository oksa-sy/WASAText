import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{path: '/home', component: HomeView},
		{path: '/session', component: LoginView}
	]
})
//controllo se l'utente vuole accedere alla home, se ha l'id allora lo faccio passare senno lo rimando alla pagina di login
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem("id");
  if (to.path !== '/session' && !token) {
    next('/session');
  } else {	
    next();
  }
});

export default router
