<template>
  <v-app>
    <v-main>
      <v-container>
        <v-alert v-if="error" color="red">Error: {{ error }}</v-alert>
        <v-row>
          <v-col v-for="(server, index) in servers" v-bind:key="index">
            <ServerComponent :server="server" @clicked="() => {servers[index].clicked = true; updateData();}"/>
          </v-col>
        </v-row>
      </v-container>
    </v-main>
  </v-app>
</template>

<script>
import ServerComponent from './components/Server';

export default {
  name: 'App',

  components: {
    ServerComponent,
  },

  data: () => ({
    error: "",
    servers: []
    // servers: [{"name":"rio","instance_state":"running","health_check_state":"initializing","tags":{"Name":"rio","aws:elasticfilesystem:default-backup":"enabled"},"ip":"34.255.136.38"},{"name":"pim","instance_state":"NONE","health_check_state":"NONE","tags":{"Name":"pim","aws:elasticfilesystem:default-backup":"enabled"},"ip":""}]
  }),
  methods: {
    updateData() {
      console.log("updating data")
      fetch("${servers_list}/")
          .then((response) => {
            if (!response.ok) {
              this.error = "failed to fetch data"
              return []
            }
            this.error = ""
            return response.json()
          })
          .then((data) => {
            if (data) {
              this.servers = data.map(server => {server.clicked=false; return server})
              this.servers.sort((a, b) => {
                a.name.localeCompare(b.name)
              })
            }
          })
    }
  },
  async created() {
    this.updateData();
    setInterval(this.updateData.bind(this), 10000)
  },
};
</script>
