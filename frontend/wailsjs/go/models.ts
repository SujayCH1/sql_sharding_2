export namespace repository {
	
	export class Project {
	    id: string;
	    name: string;
	    description: string;
	    shard_count: number;
	    status: boolean;
	    created_at: string;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.shard_count = source["shard_count"];
	        this.status = source["status"];
	        this.created_at = source["created_at"];
	    }
	}

}

