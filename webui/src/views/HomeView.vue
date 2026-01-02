<script>
export default {
	data: function() {
		return {
			errormsg: null,
			loading: false,
			some_data: null,
			search_in: null,
			search_data: null,
			show_profile: false,
			//info utente loggato
			new_username: null, //nuovo nome utente scelto
			current_username: null,
			user_photo: null,
			//conversazioni
			conv_open: false,
			conv_id: null,
			participant_id: null,
			participant_name: null,
			participant_photo: null,
			conversations: null,
			is_group: false,
			//messaggi
			msgtext: null,
			messages: null,
			selected_reaction: null,
			reply_to: null,
			msg_forward: null,
			//gruppi
			new_group: false,
			members: [],
			step: 1,
			group_name: null,
			edit_profile: false,
			group_photo: null,
			group_new_name: null,
			add_to_group: false,
		}
	},
	methods: {
		addGroupMembers(user){
			// se gia c'è lo tolgo (deselezione)
			if (this.members.includes(user)) {
				this.members = this.members.filter(id => id !== user);
			}else {
				// altrimenti lo aggiungo (selezione)
				this.members.push(user);
			}
		},
		closeProfile() {
        	this.show_profile = false;
			this.new_username = null;
    	},
		logout() {
			localStorage.removeItem("id");
			localStorage.removeItem("username");
			localStorage.removeItem("photo");
			this.$router.push("/session");
    	},
		//ricerca utenti
		async search() {
			this.errormsg = null;
			try {
				//effettua chiamata al server autenticandomi e inviando il nome letto
                let token = localStorage.getItem("id");
                let response = await this.$axios.get("/users", {
					params: {
						username: this.search_in
					},
                    headers: {
                        "Authorization": "Bearer " + token
                    }
                }); 
			this.search_data = response.data
			// elimino l'utente loggato dai risultati di ricerca 
			this.search_data = this.search_data.filter(user => user.identifier != token);
			// se sto cercando di inoltrare un messaggio includo nei risultati della ricerca anche i gruppi che corrispondono a quel nome
			if (this.msg_forward) {
				this.conv_groups = [];
				let conv_groups = this.conversations
					.filter(conv => conv.isGroup && conv.name.toLowerCase().includes(this.search_in.toLowerCase())).map(conv => ({
						identifier: conv.id, 
						username: conv.name, 
						photo: conv.photo,}));
				// aggiungo i gruppi al risultato della ricerca 
				for (const c of conv_groups){
					this.search_data.push(c);
				}
			}
			
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async changePhoto(){
			this.errormsg = null;
			let file = this.$refs.photoInput.files[0];
			if (!file) {
				return
			}
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
				let formData = new FormData();
                formData.append("photoFile", file); 
                let response = await this.$axios.put("/users/me/photo",
            		formData,
					{
						headers: {
							"Authorization": "Bearer " + token,
							"Content-Type": "multipart/form-data"
						}
					}
        		); 
			this.user_photo = response.data.photo;
			localStorage.setItem("photo", this.user_photo);
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async changeName(){
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.put("/users/me/name",
            		{name: this.new_username},
					{
						headers: {
							"Authorization": "Bearer " + token
						}
					}
        		); 
				this.current_username = this.new_username;
				localStorage.setItem("username", this.new_username);
				this.new_username = null;
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		//penso di aver bisogno dell'id della convo pero boh, l'errore è errnorows
		async startChat(user) {
			this.errormsg = null;
			this.search_data = null; 
    		this.search_in = null;  
			this.participant_id = user.identifier;
			this.participant_name = user.username;
			this.participant_photo = user.photo;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.post("/conversations",
            		{participantId: this.participant_id},
					{
						headers: {
							"Authorization": "Bearer " + token
						}
					}
        		); 
			this.conv_open = true;  
			this.conv_id = response.data.id   
			await this.listMessages(this.conv_id);  
			await this.listConversations();                                                                                           
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async openChat(conv) {
			// preparo la conversazione e chiamo list messages per preparare i messaggi
			this.conv_id = conv.id;
			this.participant_name = conv.name;
			this.participant_photo = conv.photo;
			this.is_group = conv.isGroup;
			this.conv_open = true;
			this.msgtext = null;
			await this.listMessages(conv.id);
			await this.listConversations();
		},
		async sendMessage(){
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
				if (this.reply_to === null) {
					let response = await this.$axios.post("/conversations/" + this.conv_id + "/messages",
						{content: this.msgtext},
						{
							headers: {
								"Authorization": "Bearer " + token
							}
						}); 
				}
				else {
					let response = await this.$axios.post("/conversations/" + this.conv_id + "/messages",
						{content: this.msgtext, replyTo: this.reply_to.msg_id},
						{
							headers: {
								"Authorization": "Bearer " + token
							}
						}); 
				}
			this.reply_to = null;
			this.msgtext = null;
			await this.listMessages(this.conv_id);
			await this.listConversations();
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async listMessages(conversation_id){
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.get("/conversations/" + conversation_id + "/messages",
					{
						headers: {
							"Authorization": "Bearer " + token
						}
					}
        		); 
			this.messages = response.data;
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async listConversations(){
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.get("/conversations",
					{
						headers: {
							"Authorization": "Bearer " + token
						}
					}
        		); 
			this.conversations = response.data;
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async deleteMsg(msg_id){
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.delete("/messages/" + msg_id,
					{
						headers: {
							"Authorization": "Bearer " + token
						}
					}
        		); 
			//elimino il messaggio dalla lista locale
			this.messages = this.messages.filter(msg => msg.id !== msg_id);
			await this.listMessages(this.conv_id);
			await this.listConversations();
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async reaction(msg_id, emoji){
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.post("/messages/" + msg_id + "/reaction",
					{emoji: emoji},
					{
						headers: {
							"Authorization": "Bearer " + token
						}
					}
        		); 
			await this.listMessages(this.conv_id);
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async removeReactions(msg_id, emoji){
			this.delete_reaction = false;
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.delete("/messages/" + msg_id + "/reaction",
					{
						headers: {
							"Authorization": "Bearer " + token
						},
						params: {emoji: emoji}
					}); 
			// elimino la reazione dalla vista locale
			let msg = this.messages.find(m => m.id === msg_id);
			msg.reaction = msg.reaction.filter(r => r.emoji !== emoji);	
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async forward(user_id){
			this.search_data = null; 
    		this.search_in = null;  
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.post("/messages/" + this.msg_forward,
					{id: user_id},
						{headers: {
							"Authorization": "Bearer " + token
						},
					}); 
			this.msg_forward = null;
			await this.listMessages(this.conv_id);
			await this.listConversations();
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async newGroup(){ 
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.post("/groups",
					{name: this.group_name, members: this.members},
						{headers: {
							"Authorization": "Bearer " + token
						},
					}); 
			this.new_group = false;
			this.step = 1;
			this.members = [];
			this.group_name = null;
			// per vedere se il gruppo è effettivamente un gruppo devo cercalo come conversazione, dove è effettivamente memorizzata come informazione
			// aggiorno la lista delle conversazioni 
			await this.listConversations();
			// cerco la conversazione attuale
        	let chat = this.conversations.find(conv => conv.id === response.data.id);
			this.openChat(chat);
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async leave(){
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.delete("/groups/" + this.conv_id + "/members/me",
					{
						headers: {
							"Authorization": "Bearer " + token
						}
					}
        		); 
			//elimino il messaggio dalla lista locale
			this.conversations = this.conversations.filter(conv => conv.id !== this.conv_id);
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async changeGroupPhoto(){
			this.errormsg = null;
			let file = this.$refs.groupPhotoInput.files[0];
			if (!file) {
				return
			}
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
				let formData = new FormData();
                formData.append("photoFile", file); 
                let response = await this.$axios.put("/groups/" + this.conv_id + "/photo",
            		formData,
					{
						headers: {
							"Authorization": "Bearer " + token,
							"Content-Type": "multipart/form-data"
						}
					}
        		); 
			this.group_photo = null;
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async changeGroupName(){
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.put("/groups/" + this.conv_id + "/name",
            		{name: this.group_new_name},
					{
						headers: {
							"Authorization": "Bearer " + token
						}
					}
        		); 
				this.group_new_name = null;
				this.edit_profile = false;
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		},
		async addToGroup(userId){
			this.errormsg = null;
			try {
				//effettua chiamata al server
                let token = localStorage.getItem("id");
                let response = await this.$axios.post("/groups/" + this.conv_id  + "/members",
            		{userId: userId},
					{
						headers: {
							"Authorization": "Bearer " + token
						}
					}
        		); 
				this.add_to_group = false;
				this.search_data = null; 
				this.search_in = null;  
			//controllo errori
			}catch (e) {
				if (e.response.data.message) {
                    this.errormsg = e.response.data.message; 
                } else {
                    this.errormsg = e.response.data;
                }   
			}
		}
	},
	mounted() {
		let username = localStorage.getItem("username");
		if (username) {
			this.current_username = username;
		}
		let photo = localStorage.getItem("photo");
    	if (photo) {
        	this.user_photo = photo;
    	}
		this.listConversations();
	}
}
</script>

<template>
	<div class="container">
		<!-- * barra fissa in alto allo schermo -->
		<nav class="navbar navbar-dark bg-dark fixed-top shadow">
    		<div class="container-fluid justify-content-between">
				<!-- nome  -->
				<span class="navbar-brand mb-0 h1 fw-bold bg-transparent me-3">
					WASAText 💬
				</span>
				<div class="d-flex gap-2 position-relative">
					<!-- * ricerca utenti -->
					<input v-if="!msg_forward && !new_group && !add_to_group" type="text" placeholder="Search User.." v-model="search_in">
					<button v-if="!msg_forward  && !new_group && !add_to_group" type="button" class="btn btn-sm btn-outline-primary" @click="search">
						➜
					</button>
					<!-- pulsante per il profilo modale -->
					<img v-if="user_photo" :src="'http://localhost:3000' + user_photo" class="rounded-circle shadow-sm" @click="show_profile = true" style="width: 45px; height: 45px; object-fit: cover; cursor: pointer;">
					<button v-else type="button" class="btn bg-secondary rounded-circle text-white d-flex justify-content-center align-items-center shadow-sm" ref="profileButton" style="width: 45px; height: 45px;" @click="show_profile = true">
					</button>	
					<!-- * popup si attiva se è stato schiacciato il pulsante -->
					<div v-if="show_profile" class="modal fade show" style="display: block;" aria-modal="true" role="dialog">
					<!-- il popup contiene opzioni per cambiare foto profilo e nome utente -->
						<div class="modal-dialog modal-dialog-centered">
							<div class="modal-content">
								<div class="modal-header">
									<h4 class="modal-title">Profile</h4>
									<button type="button" class="btn-close" @click="closeProfile"></button>
								</div>
								<div class="modal-body d-flex gap-2 position-relative">
									<!-- sezione foto e pulsante per cambiarla -->
									<div class="d-flex flex-column gap-2 text-start">
										<!-- deve recuperare il nome del file scelto come input -->	
										<img v-if="user_photo" :src="'http://localhost:3000' + user_photo" class="rounded-circle shadow-sm" style="width: 80px; height: 80px; object-fit: cover;">
										<span v-else class="bg-secondary rounded-circle d-flex justify-content-center align-items-center shadow-sm" style="width: 80px; height: 80px;">
										</span>
										<input type="file" accept="image/*" ref="photoInput" style="display: none" @change="changePhoto">
										<button type="button" class="btn btn-secondary btn-sm" @click="$refs.photoInput.click()">Upload</button>
									</div>
									<!-- sezione nome e input per cambiarlo -->
									<div class="d-flex flex-column gap-2 text-start">
										<h5>{{ current_username }}</h5>
										<input type="text" class="form-control" placeholder="new username..." v-model="new_username">
									</div>
								</div>
								<div class="modal-footer">	
									<button type="button" class="btn btn-primary" @click="changeName">Save username</button>
								</div>
							</div>
						</div>
					</div>
					<button class="btn btn-danger btn-sm" @click="logout">
						Logout
					</button>
					<!-- * lista degli utenti sotto la barra di ricerca -->
					<div v-if="!msg_forward  && !new_group && !add_to_group" class="list-group position-absolute shadow w-100 start-0 search-list" style="max-height: 300px; overflow-y: auto;">
						<button v-for="user in search_data" :key="user.identifier" class="list-group-item list-group-item-action d-flex align-items-center" @click="startChat(user)">
							<img v-if="user.photo" :src="'http://localhost:3000' + user.photo" class="rounded-circle shadow-sm me-3"  style="width: 40px; height: 40px;">
							<div v-else class="bg-secondary rounded-circle text-white d-flex justify-content-center align-items-center me-3" 
								style="width: 40px; height: 40px;"></div>
							<span class="fw-bold">
								{{user.username}}
							</span>
						</button>
					</div>
				</div>
    		</div> 
		</nav>
		<!-- * messaggio di errore -->
		 <div 
			v-if="errormsg" 
			class="alert alert-danger alert-dismissible fade show position-fixed top-0 start-50 translate-middle-x shadow" 
			style="z-index: 2000; margin-top: 20px; min-width: 300px;"
			role="alert">
			<strong>Error: </strong> {{ errormsg }}
			<button type="button" class="btn-close" @click="errormsg = null"></button>
		</div>
	</div>
	<!-- * schermo diviso in due colonne, una lista chat e una mostra una chat -->
		<div class="container-fluid p-0 d-flex" style="height: calc(100vh - 60px);">
			<!-- sezione lista chat  -->
			<div class="row g-0 w-100">
				<div class="col-md-4 col-lg-3 border-end h-100 d-flex flex-column">
					<div class="d-flex p-3 bg-light border-bottom align-items-center justify-content-between">
						<h3 class="me-2">Chat</h3>
						<!-- * bottone per creare un nuovo gruppo -->
						<button type="button" class="btn btn-sm btn-primary d-flex justify-content-center align-items-center"  @click="new_group=true">
							New Group
						</button>
					</div>
					<!-- lista conversazioni  -->
					<div class="flex-grow-1 overflow-auto ">
						<div  v-for="conv in conversations" :key="conv.id" class="list-group-item list-group-item-action d-flex align-items-center m-3 border-bottom" @click="openChat(conv)" style="cursor: pointer;">
							<img v-if="conv.photo" :src="'http://localhost:3000' + conv.photo" class="rounded-circle shadow-sm d-flex justify-content-center align-items-center m-1 flex-shrink-0"  style="width: 40px; height: 40px;">
							<div v-else class="bg-secondary rounded-circle d-flex justify-content-center align-items-center m-1 flex-shrink-0" 
								style="width: 40px; height: 40px;"></div>
							<div class ="d-flex flex-column overflow-hidden">
								<span class="fw-bold">
									{{conv.name}}
								</span>
								<!-- preview e data e ora di invio -->
								<div class="d-flex flex-column align-items-start w-100" style="font-size: 0.7rem;">
									<span class="text-truncate w-50 text-muted">
										{{ conv.preview }}
									</span>
									<span v-if="conv.timestamp.slice(0, 16) !== '0001-01-01T00:00'" class="opacity-75">
										{{ conv.timestamp.slice(11, 16)}} {{ conv.timestamp.slice(0, 10)}} 
									</span>
								</div>
							</div>
						</div>
					</div>
				</div>
				<!-- sezione della conversazione -->
				<div class="col-md-8 col-lg-9 h-100 d-flex flex-column bg-white">
					<!-- profilo partecipante: foto e nome -->
					<div v-if="conv_open" class="p-3 border-bottom shadow-sm d-flex align-items-center">
						<img v-if="participant_photo" :src="'http://localhost:3000' + participant_photo" class="rounded-circle shadow-sm d-flex justify-content-center align-items-center me-3 flex-shrink-0"  style="width: 50px; height: 50px;">
						<div v-else class="bg-secondary rounded-circle d-flex me-3 flex-shrink-0" 
							style="width: 50px; height: 50px;"></div>
						<h3>{{ participant_name }}</h3>
						<div v-if="is_group" class="ms-auto d-flex gap-2">
							<button type="button" class="btn btn-outline-secondary" @click="edit_profile=true">Edit</button>
							<button type="button" class="btn btn-outline-primary" @click="add_to_group=true">+ Member</button>
							<button type="button" class="btn btn-outline-danger" @click="leave">Leave</button>
						</div>
					</div>
					<!-- messaggi, se inviati dall'utente loggato sono verdi a destra altrimenti bianchi a sinistra -->
					<div class="flex-grow-1 overflow-auto p-3 d-flex flex-column" style="background-color: #e5ddd5;">
						<div class="chat-container">
							<div  v-for="msg in messages" :key="msg.id" class="mb-2 p-2 rounded shadow-sm position-relative" :class="{
									'bg-success text-white align-self-end': current_username === msg.username,
									'bg-white text-dark align-self-start': current_username !== msg.username}"
								style="max-width: 50%;">
								<span v-if="current_username !== msg.username && is_group" class="text-primary fw-bold" style="font-size: 0.80rem;;">
										{{ msg.username }}:
								</span>
								<button type="button" class="btn btn-sm dropdown-toggle position-absolute top-0 end-0 d-flex justify-content-center align-items-center text-reset" data-bs-toggle="dropdown" style="width: 10px; height: 20px; border: none;">
								</button>
								<ul class="dropdown-menu">
									<li><button class="dropdown-item" type="button" @click="reply_to={msg_id: msg.id, content: msg.content, sender: msg.username}">Reply</button></li>
									<!-- mostro l'opzione di reazione solo sui messaggi inviati dall'altro utente e uso delle reazioni preselezionate-->
									<button v-if="msg.username !== current_username" class="btn btn-sm btn-light fs-5" @click="reaction(msg.id, '❤️')">❤️</button>
									<button v-if="msg.username !== current_username" class="btn btn-sm btn-light fs-5" @click="reaction(msg.id, '😂')">😂</button>
									<button v-if="msg.username !== current_username" class="btn btn-sm btn-light fs-5" @click="reaction(msg.id, '😮')">😮</button>
									<button v-if="msg.username !== current_username" class="btn btn-sm btn-light fs-5" @click="reaction(msg.id, '😢')">😢</button>
									<button v-if="msg.username !== current_username" class="btn btn-sm btn-light fs-5" @click="reaction(msg.id, '👍')">👍</button>
									<button v-if="msg.username !== current_username" class="btn btn-sm btn-light fs-5" @click="reaction(msg.id, '🙏')">🙏</button>
									<!-- mostro l'opzione di cancellatura solo sui messaggi inviati dall'utente loggato -->
									<li v-if="msg.username === current_username"><button class="dropdown-item" @click="deleteMsg(msg.id)" type="button">Delete</button></li>
									<li><button class="dropdown-item" type="button" @click="msg_forward=msg.id">Forward</button></li>
								</ul>
								<!-- se il messaggio è in risposta a un altro messaggio lo mostro e anche da chi è stato inviato -->
								<div v-if="msg.replyToId" class="p-2 bg-white border-start border-5 border-primary rounded d-flex justify-content-between align-items-center">
								<div class="overflow-hidden">
									<span class="text-primary fw-bold" style="font-size: 0.80rem;;">
										{{ msg.username }}
									</span>
									<span class="text-truncate d-block text-muted" style="max-width: 1000px; font-size: 0.9rem;">
										{{ msg.replyToContent }}
									</span>
								</div>
							</div>
								{{ msg.content }}
								<!-- data e ora di invio -->
								<div class="d-flex justify-content-end">
									<div class="text-end opacity-75 me-3" style="font-size: 0.7rem;">
										{{ msg.timestamp.slice(11, 16)}} {{ msg.timestamp.slice(0, 10)}} 
									</div>
									<!-- una spunta ricevuto, due letto, lo mostro solo all'utente loggato -->
									<div v-if="current_username == msg.username">
										<span v-if="msg.status === 'read'" class="text-reset">✓✓</span>
										<span v-else class="text-reset">✓</span>
									</div>
									<!-- sulla stessa riga metto le reazioni -->
									<div v-if="msg.reaction" class="d-flex gap-1 mt-1 flex-wrap">
										<!-- se clicco sulla reazione, memorizzo la reazione e il messaggio, questo solo se la reazione era stata inserita dall'utente che ci clicca sopra -->
										<span v-for="reaction in msg.reaction" :key="reaction.emoji" class="badge bg-light border rounded-pill shadow-sm d-flex align-items-center" @click="reaction.username === current_username ? selected_reaction = {msg_id: msg.id, emoji: reaction.emoji} : null" style="font-size: 0.85rem; padding: 4px 4px;cursor: pointer;">
											{{ reaction.emoji }} 
											<!-- se c'e un messaggio e reazione selezionate mostro l'opzione per eliminare la reazione o una x per annullare -->
											<div v-if="selected_reaction && selected_reaction.msg_id === msg.id && selected_reaction.emoji === reaction.emoji" class="d-flex gap-1">
												<button v-if="reaction.username === current_username" type="button" class="btn btn-sm btn-danger" @click.stop="removeReactions(msg.id, reaction.emoji); selected_reaction = null">
													Delete
												</button>
												<button v-if="reaction.username === current_username" type="button" class="btn-close" @click.stop="selected_reaction = null"></button>
											</div>
										</span>
									</div>
								</div>
							</div>
						</div>
					</div>
					<!-- se è stato premuto Forward si attiva una finestra modale che permette di cercare l'utente a cui inoltrare il messaggio -->
					<div v-if="msg_forward" class="modal show" style="display: block;" tabindex="-1" role="dialog">
						<div class="modal-dialog modal-dialog-centered">
							<div class="modal-content">
								<div class="modal-header">
									<h4 class="modal-title me-3">Forward to:</h4>
									<!-- input ricerca -->
									<input type="text" placeholder="Search User.." v-model="search_in">
									<button type="button" class="btn btn-sm btn-outline-primary me-3" @click="search">
										➜
									</button>
									<button type="button" class="btn-close" @click="msg_forward=null; search_in=null; search_data = null;"></button>
								</div>
								<!-- risultati ricerca, se premi un dei risultati inoltra il messaggio a quell'utente -->
								<div class="modal-body d-flex">
									<div class="list-group w-100" style="max-height: 300px; overflow-y: auto;">
										<button v-for="user in search_data" :key="user.identifier" class="list-group-item list-group-item-action d-flex align-items-center" @click="forward(user.identifier)">
											<img v-if="user.photo" :src="'http://localhost:3000' + user.photo" class="rounded-circle shadow-sm me-3"  style="width: 40px; height: 40px;">
											<div v-else class="bg-secondary rounded-circle text-white d-flex justify-content-center align-items-center me-3" 
												style="width: 40px; height: 40px;"></div>
											<span class="fw-bold">
												{{user.username}}
											</span>
										</button>
									</div>
								</div>
							</div>
						</div>
					</div>
					<!-- aggiungere un membro al gruppo -->
					<div v-if="add_to_group" class="modal show" style="display: block;" tabindex="-1" role="dialog">
						<div class="modal-dialog modal-dialog-centered">
							<div class="modal-content">
								<div class="modal-header">
									<!-- input ricerca -->
									<h5 class="m-3">Add Member: </h5>
									<input type="text" placeholder="Search User.." v-model="search_in">
									<button type="button" class="btn btn-sm btn-outline-primary me-3" @click="search">
										➜
									</button>
									<button type="button" class="btn-close" @click="add_to_group=false; search_in=null; search_data = null;"></button>
								</div>
								<!-- risultati ricerca, se premi uno dei risultati lo aggiungo al gruppo -->	
								<div class="modal-body d-flex flex-column">
									<div class="list-group w-100" style="max-height: 300px; overflow-y: auto;">
										<button type="button" v-for="user in search_data" :key="user.identifier" class="list-group-item list-group-item-action d-flex align-items-center" @click="addToGroup(user.identifier)">
											<img v-if="user.photo" :src="'http://localhost:3000' + user.photo" class="rounded-circle shadow-sm"  style="width: 40px; height: 40px;">
											<div v-else class="bg-secondary rounded-circle text-white d-flex justify-content-center align-items-center me-3" 
												style="width: 40px; height: 40px;"></div>
											<span class="fw-bold">
												{{user.username}}
											</span>
										</button>
									</div>
								</div>
							</div>
						</div>
					</div>
					<!-- modificare il profilo del gruppo -->
					<div v-if="edit_profile" class="modal show" style="display: block;" tabindex="-1" role="dialog">
						<div class="modal-dialog modal-dialog-centered">
							<div class="modal-content">
								<div class="modal-header">
									<h4 class="modal-title">Edit group profile</h4>
									<button type="button" class="btn-close" @click="edit_profile=false"></button>
								</div>
								<div class="modal-body d-flex gap-2 position-relative">
									<!-- sezione foto e pulsante per cambiarla -->
									<div class="d-flex flex-column gap-2 text-start">
										<!-- deve recuperare il nome del file scelto come input -->	
										<img v-if="participant_photo" :src="'http://localhost:3000' + participant_photo" class="rounded-circle shadow-sm" style="width: 80px; height: 80px; object-fit: cover;">
										<span v-else class="bg-secondary rounded-circle d-flex justify-content-center align-items-center shadow-sm" style="width: 80px; height: 80px;"></span>
										<input type="file" accept="image/*" ref="groupPhotoInput" style="display: none" @change="changeGroupPhoto">
										<button type="button" class="btn btn-secondary btn-sm" @click="$refs.groupPhotoInput.click()">Upload</button>
									</div>
									<!-- sezione nome e input per cambiarlo -->
									<div class="d-flex flex-column gap-2 text-start">
										<h5>{{ group_name }}</h5>
										<input type="text" class="form-control" placeholder="new group name.." v-model="group_new_name">
									</div>
								</div>
								<div class="modal-footer">	
									<button type="button" class="btn btn-primary" @click="changeGroupName">Save</button>
								</div>
							</div>
						</div>
					</div>
					<!-- se viene premuto il pulsate per nuovo gruppo permetti di selezionare gli utenti da aggiungere -->
					<!-- controllo in quale step sono, se 1 devo mostrare le opzioni per permettere di scegliere i membri del gruppo, se due mostro le opzioni per scegliere foto e nome gruppo -->
					<div v-if="new_group" class="modal show" style="display: block;" tabindex="-1" role="dialog">
						<div class="modal-dialog modal-dialog-centered">
							<div class="modal-content">
								<div class="modal-header">
									<h4 v-if="step === 1" class="modal-title me-3">Select Members:</h4>
									<h4 v-if="step === 2" class="modal-title me-3">Group Info</h4>
									<!-- input ricerca -->
									<div v-if="step === 1">
										<input type="text" placeholder="Search User.." v-model="search_in">
										<button type="button" class="btn btn-sm btn-outline-primary me-3" @click="search">
											➜
										</button>
									</div>
									<button type="button" class="btn-close" @click="new_group=false; search_in=null; search_data = null; members=[]; step=1"></button>
								</div>
								<!-- risultati ricerca, se premi uno dei risultati aggiunge quell'utente se non c'è gia altrimenti lo toglie, mostra uno stato active se selezionato -->	
								<div class="modal-body d-flex flex-column">
									<div v-if="step === 1" class="list-group w-100" style="max-height: 300px; overflow-y: auto;">
										<button type="button" v-for="user in search_data" :key="user.identifier" class="list-group-item list-group-item-action d-flex align-items-center" :class="{active: members.includes(user.identifier)}" @click="addGroupMembers(user.identifier)">
											<img v-if="user.photo" :src="'http://localhost:3000' + user.photo" class="rounded-circle shadow-sm"  style="width: 40px; height: 40px;">
											<div v-else class="bg-secondary rounded-circle text-white d-flex justify-content-center align-items-center me-3" 
												style="width: 40px; height: 40px;"></div>
											<span class="fw-bold">
												{{user.username}}
											</span>
										</button>
									</div>
									<button v-if="members.length > 1 && step === 1" type="button" class="btn btn-success" @click="step=2; search_data = null; search_in = null;">Next</button>
									<div v-if="step===2" class="d-flex flex-column justify-content-center align-items-center">
										<input type="text" class="form-control m-2" placeholder="group name..." v-model="group_name">
										<button type="button" class="btn btn-success" @click="newGroup">Create Group</button>
									</div>
								</div>
							</div>
						</div>
					</div>
					<div class="p-3 bg-light border-top d-flex flex-column">
						<!-- se è stato selezionato un messaggio a cui rispondere lo mostro sopra l'input di testo -->
						<div v-if="reply_to" class="p-2 bg-white border-start border-5 border-primary rounded d-flex justify-content-between align-items-center">
                            <div class="overflow-hidden">
								<span class="text-primary fw-bold" style="font-size: 0.80rem;;">
                                    {{ reply_to.sender }}
                                </span>
                                <span class="text-truncate d-block text-muted" style="max-width: 1000px; font-size: 0.9rem;">
                                    {{ reply_to.content }}
                                </span>
                            </div>
                            <button type="button" class="btn-close" @click="reply_to = null"></button>
                        </div>
						<div class="p-2 bg-light border-top d-flex">
							<input v-if="conv_open" type="text" class="form-control me-1" placeholder="Write a message..." v-model="msgtext">
							<button v-if="conv_open" type="button" class="btn btn-sm btn-success" @click="sendMessage">
								➜
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>
</template>

<style>
	.container {
		 padding-top: 60px; /* spazio sulla parte anteriore */
	}
	.chat-container {
		display: flex;
		flex-direction: column-reverse; /* inverte lo scorrimento, dal basso verso l'alto */
		overflow-y: auto;
		height: 600px;
	}
	.search-list {
		top: 130%;             /* porta sotto l'input */
        z-index: 1050;         /* sopra il resto */
        max-height: 300px;     /* lunghezza massima della tendina */
        overflow-y: auto;      /* scrollbar quando supera la lunghezza massima */
	}
</style>

<!-- ! controllare: 
visualizzazione messaggi gruppo
per vedere la foto del profilo bisogna fare refresh  
fare i linter -->