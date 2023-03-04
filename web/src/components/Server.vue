<template>
  <v-container>
    <v-card
        elevation="2"
        :loading="!stateStable()"
    >
      <v-card-title>{{ server.name }}.${domain_name}</v-card-title>
      <v-card-text>
        <v-btn
            v-if="startable()"
            @click="start_server()"
            color="primary"
            :loading="server.clicked"
        >Start</v-btn>
        <v-btn
            v-else
            @click="stop_server()"
            color="primary"
            :loading="server.clicked"
        >Stop</v-btn>
        <v-list>
          <v-list-item v-if="server.ip!==''">
            <v-list-item-title>IP</v-list-item-title>
            <v-list-item-subtitle>{{server.ip}}</v-list-item-subtitle>
          </v-list-item>
          <v-list-item>
            <v-list-item-title>Status</v-list-item-title>
            <v-tooltip bottom>
              <template v-slot:activator="{ on, attrs }">
                <v-list-item-subtitle v-bind="attrs" v-on="on">
                  <!-- instance state -->
                  <v-icon v-if="server.instance_state==='running'" color="green">mdi-checkbox-marked-circle-outline</v-icon>
                  <v-icon v-else-if="startable()" color="red">mdi-close-circle-outline</v-icon>
                  <v-progress-circular
                      v-else
                      :color="server.instance_state==='shutting-down'?'red':'green'"
                      indeterminate
                    ></v-progress-circular>
                  <!-- health check state -->
                  <v-icon v-if="server.health_check_state==='ok'" color="green">mdi-checkbox-marked-circle-outline</v-icon>
                  <v-icon v-else-if="server.health_check_state==='NONE'" color="red">mdi-close-circle-outline</v-icon>
                  <v-progress-circular
                      v-else
                      :color="server.health_check_state==='initializing'?'green':'red'"
                      indeterminate
                    ></v-progress-circular>
                </v-list-item-subtitle>
              </template>
              <span>{{ server.instance_state }} / {{ server.health_check_state }}</span>
            </v-tooltip>
          </v-list-item>
        </v-list>
        <v-alert>{{ error }}</v-alert>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script>
  export default {
    name: 'ServerComponent',
    props: ["server"],
    data: () => ({
      error: "",
      // wantedState: "",
      // setState: "",
    }),
    methods: {
      start_server() {
        // this.wantedState = "START"
        fetch("${server_stop}/?name="+this.server.name, { method: "PUT" })
            .then((response) => {
              if (!response.ok) {
                return response.text()
              }
            })
            .then((text) => {
              this.error = text
            })
        this.$emit('clicked', '')
      },
      stop_server() {
        // this.wantedState = "STOP"
        fetch("${server_start}/?name="+this.server.name, { method: "DELETE" })
            .then((response) => {
              if (!response.ok) {
                return response.text()
              }
            })
            .then((text) => {
              this.error = text
            })
        this.$emit('clicked', '')
      },
      stateStable() {
        return this.server.health_check_state === "ok" || ['NONE', 'terminated', 'running'].includes(this.server.instance_state)
      },
      startable() {
        return ['terminated', 'NONE'].includes(this.server.instance_state)
      },
        /*
         * instance states:
         * pending
         * running
         * shutting-down
         * terminated
         * stopping
         * stopped
         *
         * healthcheck:
         * ok
         * impaired
         * insufficient-data
         * not-applicable
         * initializing
         */
    },
    // async created() {
    //   this.setState = this.wantedState = this.server.instance_state === "NONE"?"STOP":"START"
    // },
    // async updated() {
    //   this.setState = this.server.instance_state === "NONE"?"STOP":"START"
    // },

  }
</script>
