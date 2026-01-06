<!-- pagina di login, prima schermata mostrata a un utente non loggato -->
<script>
export default {
	data: function() {
		return {
			username: null,
            errormsg: null
		}
	},
	methods: {
		async login() {
			try {
                //chiamata al server
				let response = await this.$axios.post("/session", {name: this.username});
                //salvo il token id dato come risposta
			    localStorage.setItem("id", response.data.identifier);
                localStorage.setItem("username", this.username);
                //passo alla home
                this.$router.push("/home");
            } catch (e) {
                //controllo se c'è un errore, se presente un messaggio stampo quello, altrimenti l'errore in stringa
                if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }                                      
            }
        },
	}
}           
</script>

<template>
  <div>
    <nav class="navbar navbar-dark bg-dark fixed-top shadow">
      <div class="container-fluid">
        <span class="navbar-brand mb-0 h1 fw-bold bg-transparent">
          WASAText 💬
        </span>
      </div>
    </nav>
    <div class="d-flex justify-content-center align-items-center vh-100">
      <div class="card p-4 shadow">
        <h2 class="text-center mb-4">Login</h2>
        <div class="mb-3">
          <input v-model="username" type="text" class="form-control" placeholder="username">
          <div v-if="errormsg" class="alert alert-danger">{{ errormsg }}</div>
        </div>
        <button class="btn btn-primary w-100" @click="login">Sign In</button>
      </div>
    </div>
  </div>
</template>

<style>
</style>