import { Container } from "@cloudflare/containers";

export class MiAction extends Container {
  defaultPort = 8080;
  sleepAfter = "10m";

  constructor(ctx, env, options) {
    super(ctx, env, options);
    this.envVars = {
      LEGISCAN_API_KEY: env.LEGISCAN_API_KEY,
    };
  }
}

export default {
  async fetch(request, env) {
    const container = env.MIACTION.getByName("main");
    return container.fetch(request);
  },
};
