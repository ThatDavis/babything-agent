export namespace agent {
	
	export class Config {
	    cloud_url: string;
	    rtsp_url: string;
	    agent_token: string;
	    turn_url?: string;
	    turn_username?: string;
	    turn_password?: string;
	    log_level?: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cloud_url = source["cloud_url"];
	        this.rtsp_url = source["rtsp_url"];
	        this.agent_token = source["agent_token"];
	        this.turn_url = source["turn_url"];
	        this.turn_username = source["turn_username"];
	        this.turn_password = source["turn_password"];
	        this.log_level = source["log_level"];
	    }
	}

}

