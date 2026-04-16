export namespace app {
	
	export class BuildingTypeInfo {
	    type: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildingTypeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	    }
	}
	export class LogEntry {
	    message: string;
	    level: string;
	    time: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.level = source["level"];
	        this.time = source["time"];
	    }
	}
	export class TaskItem {
	    task: string;
	    executeAt: string;
	    stage: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.task = source["task"];
	        this.executeAt = source["executeAt"];
	        this.stage = source["stage"];
	    }
	}
	export class TroopTypeInfo {
	    type: number;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new TroopTypeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	    }
	}

}

export namespace database {
	
	export class AccessDetail {
	    id: number;
	    username: string;
	    password: string;
	    proxyHost: string;
	    proxyPort: number;
	    proxyUsername: string;
	    proxyPassword: string;
	    useragent: string;
	    lastUsed: string;
	
	    static createFrom(source: any = {}) {
	        return new AccessDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.proxyHost = source["proxyHost"];
	        this.proxyPort = source["proxyPort"];
	        this.proxyUsername = source["proxyUsername"];
	        this.proxyPassword = source["proxyPassword"];
	        this.useragent = source["useragent"];
	        this.lastUsed = source["lastUsed"];
	    }
	}
	export class AccountDetail {
	    id: number;
	    username: string;
	    server: string;
	    accesses: AccessDetail[];
	
	    static createFrom(source: any = {}) {
	        return new AccountDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.server = source["server"];
	        this.accesses = this.convertValues(source["accesses"], AccessDetail);
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
	export class AccountListItem {
	    id: number;
	    username: string;
	    server: string;
	
	    static createFrom(source: any = {}) {
	        return new AccountListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.server = source["server"];
	    }
	}
	export class BuildingItem {
	    id: number;
	    type: number;
	    typeName: string;
	    level: number;
	    maxLevel: number;
	    isUnderConstruction: boolean;
	    location: number;
	    color: string;
	
	    static createFrom(source: any = {}) {
	        return new BuildingItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.typeName = source["typeName"];
	        this.level = source["level"];
	        this.maxLevel = source["maxLevel"];
	        this.isUnderConstruction = source["isUnderConstruction"];
	        this.location = source["location"];
	        this.color = source["color"];
	    }
	}
	export class FarmItem {
	    id: number;
	    name: string;
	    isActive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new FarmItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.isActive = source["isActive"];
	    }
	}
	export class JobItem {
	    id: number;
	    position: number;
	    type: number;
	    content: string;
	    display: string;
	
	    static createFrom(source: any = {}) {
	        return new JobItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.position = source["position"];
	        this.type = source["type"];
	        this.content = source["content"];
	        this.display = source["display"];
	    }
	}
	export class NormalBuildInput {
	    villageId: number;
	    type: number;
	    level: number;
	    location: number;
	
	    static createFrom(source: any = {}) {
	        return new NormalBuildInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.villageId = source["villageId"];
	        this.type = source["type"];
	        this.level = source["level"];
	        this.location = source["location"];
	    }
	}
	export class QueueBuildingItem {
	    position: number;
	    location: number;
	    typeName: string;
	    level: number;
	    completeTime: string;
	
	    static createFrom(source: any = {}) {
	        return new QueueBuildingItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.position = source["position"];
	        this.location = source["location"];
	        this.typeName = source["typeName"];
	        this.level = source["level"];
	        this.completeTime = source["completeTime"];
	    }
	}
	export class ResourceBuildInput {
	    villageId: number;
	    plan: number;
	    level: number;
	
	    static createFrom(source: any = {}) {
	        return new ResourceBuildInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.villageId = source["villageId"];
	        this.plan = source["plan"];
	        this.level = source["level"];
	    }
	}
	export class StorageDTO {
	    wood: number;
	    clay: number;
	    iron: number;
	    crop: number;
	    warehouse: number;
	    granary: number;
	    freeCrop: number;
	
	    static createFrom(source: any = {}) {
	        return new StorageDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.wood = source["wood"];
	        this.clay = source["clay"];
	        this.iron = source["iron"];
	        this.crop = source["crop"];
	        this.warehouse = source["warehouse"];
	        this.granary = source["granary"];
	        this.freeCrop = source["freeCrop"];
	    }
	}
	export class TransferRuleDTO {
	    id: number;
	    villageId: number;
	    position: number;
	    targetVillageId: number;
	    targetName: string;
	    wood: number;
	    clay: number;
	    iron: number;
	    crop: number;
	
	    static createFrom(source: any = {}) {
	        return new TransferRuleDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.villageId = source["villageId"];
	        this.position = source["position"];
	        this.targetVillageId = source["targetVillageId"];
	        this.targetName = source["targetName"];
	        this.wood = source["wood"];
	        this.clay = source["clay"];
	        this.iron = source["iron"];
	        this.crop = source["crop"];
	    }
	}
	export class TransferRuleInput {
	    villageId: number;
	    targetVillageId: number;
	    wood: number;
	    clay: number;
	    iron: number;
	    crop: number;
	
	    static createFrom(source: any = {}) {
	        return new TransferRuleInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.villageId = source["villageId"];
	        this.targetVillageId = source["targetVillageId"];
	        this.wood = source["wood"];
	        this.clay = source["clay"];
	        this.iron = source["iron"];
	        this.crop = source["crop"];
	    }
	}
	export class VillageListItem {
	    id: number;
	    name: string;
	    x: number;
	    y: number;
	    isActive: boolean;
	    isUnderAttack: boolean;
	    evasionState: number;
	
	    static createFrom(source: any = {}) {
	        return new VillageListItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.isActive = source["isActive"];
	        this.isUnderAttack = source["isUnderAttack"];
	        this.evasionState = source["evasionState"];
	    }
	}

}

