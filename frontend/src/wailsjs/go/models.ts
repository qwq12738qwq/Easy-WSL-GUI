export namespace runtimeGUI {
	
	export class List {
	    name: string;
	    status: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new List(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.status = source["status"];
	        this.version = source["version"];
	    }
	}

}

