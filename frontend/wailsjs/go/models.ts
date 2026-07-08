export namespace installer {
	
	export class OLCRTC {
	    provider: string;
	    transport: string;
	    room: string;
	
	    static createFrom(source: any = {}) {
	        return new OLCRTC(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.transport = source["transport"];
	        this.room = source["room"];
	    }
	}
	export class Ports {
	    vless: string;
	    hysteria2: string;
	    mieru: string;
	    amneziawg: string;
	    naive: string;
	
	    static createFrom(source: any = {}) {
	        return new Ports(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vless = source["vless"];
	        this.hysteria2 = source["hysteria2"];
	        this.mieru = source["mieru"];
	        this.amneziawg = source["amneziawg"];
	        this.naive = source["naive"];
	    }
	}
	export class Protocols {
	    vless: boolean;
	    hysteria2: boolean;
	    mieru: boolean;
	    amneziawg: boolean;
	    naive: boolean;
	    olcrtc: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Protocols(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.vless = source["vless"];
	        this.hysteria2 = source["hysteria2"];
	        this.mieru = source["mieru"];
	        this.amneziawg = source["amneziawg"];
	        this.naive = source["naive"];
	        this.olcrtc = source["olcrtc"];
	    }
	}
	export class Request {
	    host: string;
	    sshPort: string;
	    user: string;
	    password: string;
	    domain: string;
	    email: string;
	    sni: string;
	    hy2Obfs: string;
	    protocols: Protocols;
	    ports: Ports;
	    olcrtc: OLCRTC;
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.sshPort = source["sshPort"];
	        this.user = source["user"];
	        this.password = source["password"];
	        this.domain = source["domain"];
	        this.email = source["email"];
	        this.sni = source["sni"];
	        this.hy2Obfs = source["hy2Obfs"];
	        this.protocols = this.convertValues(source["protocols"], Protocols);
	        this.ports = this.convertValues(source["ports"], Ports);
	        this.olcrtc = this.convertValues(source["olcrtc"], OLCRTC);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

