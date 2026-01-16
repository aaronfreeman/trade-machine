export namespace models {
	
	export class AgentRun {
	    id: number[];
	    agent_type: string;
	    symbol?: string;
	    status: string;
	    input_data?: Record<string, any>;
	    output_data?: Record<string, any>;
	    error_message?: string;
	    duration_ms: number;
	    // Go type: time
	    started_at: any;
	    // Go type: time
	    completed_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new AgentRun(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.agent_type = source["agent_type"];
	        this.symbol = source["symbol"];
	        this.status = source["status"];
	        this.input_data = source["input_data"];
	        this.output_data = source["output_data"];
	        this.error_message = source["error_message"];
	        this.duration_ms = source["duration_ms"];
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.completed_at = this.convertValues(source["completed_at"], null);
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
	export class Position {
	    id: number[];
	    symbol: string;
	    // Go type: decimal
	    quantity: any;
	    // Go type: decimal
	    avg_entry_price: any;
	    // Go type: decimal
	    current_price: any;
	    // Go type: decimal
	    unrealized_pl: any;
	    side: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Position(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.symbol = source["symbol"];
	        this.quantity = this.convertValues(source["quantity"], null);
	        this.avg_entry_price = this.convertValues(source["avg_entry_price"], null);
	        this.current_price = this.convertValues(source["current_price"], null);
	        this.unrealized_pl = this.convertValues(source["unrealized_pl"], null);
	        this.side = source["side"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class Recommendation {
	    id: number[];
	    symbol: string;
	    action: string;
	    // Go type: decimal
	    quantity: any;
	    // Go type: decimal
	    target_price: any;
	    confidence: number;
	    reasoning: string;
	    fundamental_score: number;
	    sentiment_score: number;
	    technical_score: number;
	    status: string;
	    // Go type: time
	    approved_at?: any;
	    // Go type: time
	    rejected_at?: any;
	    executed_trade_id?: number[];
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Recommendation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.symbol = source["symbol"];
	        this.action = source["action"];
	        this.quantity = this.convertValues(source["quantity"], null);
	        this.target_price = this.convertValues(source["target_price"], null);
	        this.confidence = source["confidence"];
	        this.reasoning = source["reasoning"];
	        this.fundamental_score = source["fundamental_score"];
	        this.sentiment_score = source["sentiment_score"];
	        this.technical_score = source["technical_score"];
	        this.status = source["status"];
	        this.approved_at = this.convertValues(source["approved_at"], null);
	        this.rejected_at = this.convertValues(source["rejected_at"], null);
	        this.executed_trade_id = source["executed_trade_id"];
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class Trade {
	    id: number[];
	    symbol: string;
	    side: string;
	    // Go type: decimal
	    quantity: any;
	    // Go type: decimal
	    price: any;
	    // Go type: decimal
	    total_value: any;
	    // Go type: decimal
	    commission: any;
	    status: string;
	    alpaca_order_id?: string;
	    // Go type: time
	    executed_at?: any;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Trade(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.symbol = source["symbol"];
	        this.side = source["side"];
	        this.quantity = this.convertValues(source["quantity"], null);
	        this.price = this.convertValues(source["price"], null);
	        this.total_value = this.convertValues(source["total_value"], null);
	        this.commission = this.convertValues(source["commission"], null);
	        this.status = source["status"];
	        this.alpaca_order_id = source["alpaca_order_id"];
	        this.executed_at = this.convertValues(source["executed_at"], null);
	        this.created_at = this.convertValues(source["created_at"], null);
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

