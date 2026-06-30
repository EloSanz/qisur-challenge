import { useState, useEffect } from 'react';
import axios from 'axios';
import { ReactFlow, MiniMap, Controls, Background, MarkerType, Position } from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { Activity, Clock, Database, Radio, CheckCircle, Zap } from 'lucide-react';

const API_URL = '/qisur/api';

type AuditTrace = {
  id: string;
  trace_id: string;
  action: string;
  entity_type: string;
  entity_id: string;
  timestamp: string;
};

// Map actions to specific icons and colors
const getActionConfig = (action: string) => {
  if (action.includes('DB_SAVED')) return { icon: <Database className="w-4 h-4 text-emerald-400" />, color: 'emerald', glow: 'shadow-[0_0_15px_rgba(52,211,153,0.3)]' };
  if (action.includes('PUBLISHED')) return { icon: <Radio className="w-4 h-4 text-yellow-400" />, color: 'yellow', glow: 'shadow-[0_0_15px_rgba(250,204,21,0.3)]' };
  if (action.includes('CONSUMED')) return { icon: <Zap className="w-4 h-4 text-amber-500" />, color: 'amber', glow: 'shadow-[0_0_15px_rgba(245,158,11,0.3)]' };
  if (action.includes('DELIVERED')) return { icon: <CheckCircle className="w-4 h-4 text-orange-400" />, color: 'orange', glow: 'shadow-[0_0_15px_rgba(251,146,60,0.3)]' };
  return { icon: <Activity className="w-4 h-4 text-gray-400" />, color: 'gray', glow: 'shadow-[0_0_10px_rgba(156,163,175,0.2)]' };
};

export default function App() {
  const [recentTraces, setRecentTraces] = useState<AuditTrace[]>([]);
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);
  const [nodes, setNodes] = useState<any[]>([]);
  const [edges, setEdges] = useState<any[]>([]);

  useEffect(() => {
    fetchRecentTraces();
    const interval = setInterval(fetchRecentTraces, 3000);
    return () => clearInterval(interval);
  }, []);

  const fetchRecentTraces = async () => {
    try {
      const res = await axios.get(`${API_URL}/traces`);
      setRecentTraces(res.data || []);
    } catch (err) {
      console.error('Error fetching traces:', err);
    }
  };

  const fetchTraceDetails = async (traceId: string) => {
    setSelectedTraceId(traceId);
    try {
      const res = await axios.get(`${API_URL}/traces/${traceId}`);
      const history: AuditTrace[] = res.data || [];
      
      const newNodes = history.map((trace, index) => {
        const config = getActionConfig(trace.action);
        return {
          id: trace.id,
          position: { x: 300, y: index * 140 + 50 },
          sourcePosition: Position.Bottom,
          targetPosition: Position.Top,
          data: { 
            label: (
              <div className={`flex flex-col p-4 rounded-xl glass-panel ${config.glow} min-w-[260px] group hover:scale-105 transition-transform duration-300`}>
                <div className="flex items-center justify-between mb-3 border-b border-white/10 pb-2">
                  <div className={`flex items-center gap-2 font-mono text-sm tracking-wider text-${config.color}-400`}>
                    {config.icon}
                    {trace.action}
                  </div>
                  <div className="flex items-center text-xs text-gray-500 bg-black/40 px-2 py-1 rounded-full border border-white/5">
                    <Clock className="w-3 h-3 mr-1 opacity-70" />
                    {new Date(trace.timestamp).toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit', fractionalSecondDigits: 3 })}
                  </div>
                </div>
                <div className="flex items-center gap-2 text-xs text-gray-400 font-mono">
                  <span className="bg-white/5 px-2 py-1 rounded uppercase tracking-wider">{trace.entity_type}</span>
                  <span className="truncate opacity-50">{trace.entity_id}</span>
                </div>
              </div>
            ) 
          },
          type: 'default',
          style: {
            background: 'transparent',
            border: 'none',
            boxShadow: 'none',
            padding: 0
          }
        };
      });

      const newEdges = history.slice(0, -1).map((trace, index) => {
        const sourceConfig = getActionConfig(trace.action);
        return {
          id: `e-${trace.id}-${history[index+1].id}`,
          source: trace.id,
          target: history[index+1].id,
          animated: true,
          style: { stroke: sourceConfig.color === 'emerald' ? '#34d399' : sourceConfig.color === 'yellow' ? '#facc15' : sourceConfig.color === 'amber' ? '#f59e0b' : '#fb923c', strokeWidth: 2, opacity: 0.6 },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: sourceConfig.color === 'emerald' ? '#34d399' : sourceConfig.color === 'yellow' ? '#facc15' : sourceConfig.color === 'amber' ? '#f59e0b' : '#fb923c',
          },
        };
      });

      setNodes(newNodes);
      setEdges(newEdges);

    } catch (err) {
      console.error('Error fetching trace details:', err);
    }
  };

  return (
    <div className="flex h-screen w-full bg-[#050505] overflow-hidden text-gray-200">
      
      {/* Sidebar */}
      <div className="w-[380px] flex-shrink-0 border-r border-white/10 glass-panel flex flex-col h-full z-20">
        
        {/* Header */}
        <div className="p-6 border-b border-white/5 bg-gradient-to-b from-white/5 to-transparent flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-yellow-400 to-amber-500 tracking-tight">
              Event Tracker
            </h1>
            <p className="text-xs text-gray-500 font-mono mt-1 flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></span>
              Live Telemetry
            </p>
          </div>
          <div className="p-2 rounded-xl bg-white/5 border border-white/10">
            <Activity className="w-5 h-5 text-yellow-400" />
          </div>
        </div>

        {/* Traces List */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {recentTraces.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-center px-4">
              <div className="w-16 h-16 rounded-2xl bg-white/5 border border-white/10 flex items-center justify-center mb-4">
                <Radio className="w-8 h-8 text-gray-600 animate-pulse-slow" />
              </div>
              <h3 className="text-sm font-medium text-gray-300">No traces detected</h3>
              <p className="text-xs text-gray-500 mt-2 max-w-[200px]">Perform an action in the API to capture live events.</p>
            </div>
          ) : (
            recentTraces.map((trace) => (
              <div 
                key={trace.trace_id}
                onClick={() => fetchTraceDetails(trace.trace_id)}
                className={`p-4 rounded-xl cursor-pointer glass-card group relative overflow-hidden ${selectedTraceId === trace.trace_id ? 'bg-yellow-500/10 border-yellow-500/30 shadow-[0_0_20px_rgba(250,204,21,0.15)]' : ''}`}
              >
                {/* Accent glow on select */}
                {selectedTraceId === trace.trace_id && (
                  <div className="absolute left-0 top-0 bottom-0 w-1 bg-yellow-400 shadow-[0_0_10px_rgba(250,204,21,0.8)]"></div>
                )}
                
                <div className="flex justify-between items-start mb-3">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-mono text-yellow-400 bg-yellow-500/10 px-2 py-0.5 rounded border border-yellow-500/20">
                      {trace.trace_id.split('-')[0]}
                    </span>
                  </div>
                  <span className="text-[11px] text-gray-500 font-mono">
                    {new Date(trace.timestamp).toLocaleTimeString([], { hour12: false })}
                  </span>
                </div>
                
                <div className="flex items-center justify-between">
                  <div className="flex flex-col">
                    <span className="text-sm font-medium text-gray-300 group-hover:text-white transition-colors">{trace.entity_type}</span>
                    <span className="text-[10px] text-gray-600 font-mono mt-0.5 truncate max-w-[200px]">{trace.entity_id}</span>
                  </div>
                  <div className="w-6 h-6 rounded-full bg-white/5 border border-white/10 flex items-center justify-center group-hover:bg-white/10 transition-colors">
                    <Zap className="w-3 h-3 text-gray-400 group-hover:text-amber-400" />
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Main Canvas */}
      <div className="flex-1 relative bg-[#050505]">
        {/* Subtle grid gradient overlay */}
        <div className="absolute inset-0 pointer-events-none bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-yellow-900/10 via-[#050505] to-[#050505] z-0"></div>

        {selectedTraceId ? (
          <ReactFlow 
            nodes={nodes} 
            edges={edges} 
            fitView 
            fitViewOptions={{ padding: 0.3 }}
            className="z-10"
            minZoom={0.5}
            maxZoom={2}
          >
            <Background color="#333" gap={24} size={2} className="opacity-40" />
            <Controls showInteractive={false} className="!bg-black/50 !border-white/10 !backdrop-blur-md rounded-xl overflow-hidden shadow-2xl" />
            <MiniMap 
              nodeColor="#facc15" 
              maskColor="rgba(0, 0, 0, 0.7)" 
              className="!bg-black/80 !border-white/10 rounded-xl overflow-hidden shadow-2xl"
            />
          </ReactFlow>
        ) : (
          <div className="absolute inset-0 flex items-center justify-center flex-col z-10">
            <div className="relative">
              <div className="absolute inset-0 bg-yellow-500/10 blur-3xl rounded-full w-32 h-32 animate-pulse-slow"></div>
              <Activity className="w-20 h-20 text-gray-800 relative z-10" strokeWidth={1} />
            </div>
            <p className="text-xl font-medium text-gray-500 mt-6 tracking-tight">Awaiting Trace Selection</p>
            <p className="text-sm text-gray-600 mt-2 max-w-sm text-center">Select an event from the sidebar to visualize its lifecycle across microservices.</p>
          </div>
        )}
      </div>
    </div>
  );
}
