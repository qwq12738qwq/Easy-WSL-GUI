export namespace main {
	
	export class MigrationOptions {
	    sourcePath: string;
	    targetPath: string;
	    distroName: string;
	
	    static createFrom(source: any = {}) {
	        return new MigrationOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourcePath = source["sourcePath"];
	        this.targetPath = source["targetPath"];
	        this.distroName = source["distroName"];
	    }
	}

}

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
	export class Metrics {
	    cpu: string;
	    memUsed: string;
	    memTotal: string;
	    usedBytes: number;
	    totalBytes: number;
	    disk: string;
	
	    static createFrom(source: any = {}) {
	        return new Metrics(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.cpu = source["cpu"];
	        this.memUsed = source["memUsed"];
	        this.memTotal = source["memTotal"];
	        this.usedBytes = source["usedBytes"];
	        this.totalBytes = source["totalBytes"];
	        this.disk = source["disk"];
	    }
	}

}

export namespace setting {
	
	export class PerformanceConfig {
	    memoryLimit: number;
	    swap: number;
	    swapFile: string;
	    processorCount: number;
	    networkMode: string;
	    localhostForwarding: boolean;
	    autoMemoryReclaim: string;
	    sparseVhd: boolean;
	    dnsTunneling: boolean;
	    firewall: boolean;
	    autoProxy: boolean;
	    hostAddressLoopback: boolean;
	    guiApplications: boolean;
	    debugConsole: boolean;
	    kernel: string;
	    kernelModules: string;
	    kernelCommandLine: string;
	    safeMode: boolean;
	    maxCrashDumpCount: number;
	    nestedVirtualization: boolean;
	    vmIdleTimeout: number;
	    dnsProxy: boolean;
	    defaultVhdSize: number;
	    pageReporting: boolean;
	    bestEffortDnsParsing: boolean;
	    dnsTunnelingIpAddress: string;
	    initialAutoProxyTimeout: number;
	    ignoredPorts: string;
	
	    static createFrom(source: any = {}) {
	        return new PerformanceConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.memoryLimit = source["memoryLimit"];
	        this.swap = source["swap"];
	        this.swapFile = source["swapFile"];
	        this.processorCount = source["processorCount"];
	        this.networkMode = source["networkMode"];
	        this.localhostForwarding = source["localhostForwarding"];
	        this.autoMemoryReclaim = source["autoMemoryReclaim"];
	        this.sparseVhd = source["sparseVhd"];
	        this.dnsTunneling = source["dnsTunneling"];
	        this.firewall = source["firewall"];
	        this.autoProxy = source["autoProxy"];
	        this.hostAddressLoopback = source["hostAddressLoopback"];
	        this.guiApplications = source["guiApplications"];
	        this.debugConsole = source["debugConsole"];
	        this.kernel = source["kernel"];
	        this.kernelModules = source["kernelModules"];
	        this.kernelCommandLine = source["kernelCommandLine"];
	        this.safeMode = source["safeMode"];
	        this.maxCrashDumpCount = source["maxCrashDumpCount"];
	        this.nestedVirtualization = source["nestedVirtualization"];
	        this.vmIdleTimeout = source["vmIdleTimeout"];
	        this.dnsProxy = source["dnsProxy"];
	        this.defaultVhdSize = source["defaultVhdSize"];
	        this.pageReporting = source["pageReporting"];
	        this.bestEffortDnsParsing = source["bestEffortDnsParsing"];
	        this.dnsTunnelingIpAddress = source["dnsTunnelingIpAddress"];
	        this.initialAutoProxyTimeout = source["initialAutoProxyTimeout"];
	        this.ignoredPorts = source["ignoredPorts"];
	    }
	}

}

