<template>
  <v-container>
    <v-card
        elevation="2"
        :loading="statusValue() > 0 && statusValue() < 100"
    >
      <v-tooltip bottom>
        <template v-slot:activator="{ on, attrs }">
        <v-card-title
            v-bind="attrs"
            v-on="on"
        >
          {{ server.name }}.${domain_name}
          <v-spacer></v-spacer>
          Status:
          <v-progress-circular
              v-if="statusValue() > 0 && statusValue() < 100"
              :value="statusValue()"
              :color="statusValue()===100?'green':statusValue()===0?'red':'orange'"
          />
          <v-icon v-if="statusValue() === 100" color="green">mdi-checkbox-marked-circle-outline</v-icon>
          <v-icon v-if="statusValue() === 0" color="red">mdi-close-circle-outline</v-icon>
        </v-card-title>
        </template>
        <span>{{ server.ip }}</span>
      </v-tooltip>
      <v-card-text>
        <v-btn
            v-if="server.last_status === 'NONE'"
            @click="start_server()"
            color="primary"
            :loading="server.clicked"
        >Start</v-btn>
        <v-btn
            v-if="server.last_status !== 'NONE'"
            @click="stop_server()"
            color="primary"
            :loading="server.clicked"
        >Stop</v-btn>
        <v-alert>{{ error }}</v-alert>
        <v-list>
          <v-list-item v-for="(value,key) in server.tags" v-bind:key="key">
            <v-list-item-title>{{ key }}</v-list-item-title>
            <v-list-item-subtitle>{{ value }}</v-list-item-subtitle>
          </v-list-item>
        </v-list>
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
      wantedState: "",
      setState: "",
    }),
    methods: {
      start_server() {
        this.wantedState = "START"
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
        this.wantedState = "STOP"
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
      statusValue() {
        if (this.wantedState !== this.setState) return 10
        if (this.server.last_status === "NONE") return 0
        if (this.server.last_status === "STOPPED") return 0
        if (this.server.last_status === "PROVISIONING") return 25
        if (this.server.last_status === "RUNNING") return 100
        if (this.server.desired_status === "STOPPED") return 50
        // last_status is pending
        if (this.server.health_status === "UNKNOWN") return 50
        if (this.server.health_status === "HEALTHY") return 75
        return 0
      }
    },
    async created() {
      this.setState = this.wantedState = this.server.last_status === "NONE"?"STOP":"START"
    },
    async updated() {
      this.setState = this.server.last_status === "NONE"?"STOP":"START"
    },

  }
</script>
