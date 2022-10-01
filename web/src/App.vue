<template>
  <v-app>
    <v-main>
      <v-container>
        <v-row>
          <v-col v-for="server in servers" v-bind:key="server.name">
            <ServerComponent :server="server" @update="updateData"/>
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
  }),
  methods: {
    updateData() {
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
              this.servers = data
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
