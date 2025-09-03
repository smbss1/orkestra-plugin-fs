package shared

import (
	"encoding/gob"
	"encoding/json"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// HandshakeConfig est utilisé pour s'assurer que le moteur et le plugin
// communiquent sur la même version.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "ORKESTRA_PLUGIN",
	MagicCookieValue: "hello",
}

// NodeExecutor est l'interface que tous les plugins de nœuds doivent implémenter.
type NodeExecutor interface {
	Execute(node Node, ctx ExecutionContext) (interface{}, error)
	GetCapabilities() ([]string, error)
}

// ExecutionContext contient toutes les données nécessaires à l'exécution d'un nœud.
// Les champs doivent être exportés (commencer par une majuscule) pour être
// accessibles par le système RPC.
type ExecutionContext struct {
	TriggerData map[string]interface{}
	NodeOutputs map[string]interface{}
	Secrets     map[string]string
	CurrentItem interface{}
	FailureData map[string]interface{}
}

// Node est une version "sérialisable" de la structure Node du moteur.
// Les champs complexes comme 'With' sont passés en JSON brut pour éviter
// des problèmes de type complexes avec gob.
type Node struct {
	ID        string
	Uses      string
	With      json.RawMessage
	Needs     []string
	Do        json.RawMessage
	Retries   json.RawMessage
	OnFailure json.RawMessage
}

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}

// --- Implémentation du wrapper go-plugin ---

// NodeExecutorRPC est l'implémentation côté client (moteur) pour appeler le plugin via RPC.
type NodeExecutorRPC struct{ client *rpc.Client }

func (g *NodeExecutorRPC) Execute(node Node, ctx ExecutionContext) (interface{}, error) {
	var resp interface{}
	args := struct {
		Node Node
		Ctx  ExecutionContext
	}{node, ctx}
	err := g.client.Call("Plugin.Execute", args, &resp)
	return resp, err
}

func (g *NodeExecutorRPC) GetCapabilities() ([]string, error) {
	var resp []string
	err := g.client.Call("Plugin.GetCapabilities", new(interface{}), &resp)
	return resp, err
}

// NodeExecutorRPCServer est l'implémentation côté serveur (plugin) qui expose les méthodes.
type NodeExecutorRPCServer struct{ Impl NodeExecutor }

func (s *NodeExecutorRPCServer) Execute(args struct {
	Node Node
	Ctx  ExecutionContext
}, resp *interface{}) error {
	var err error
	*resp, err = s.Impl.Execute(args.Node, args.Ctx)
	return err
}

func (s *NodeExecutorRPCServer) GetCapabilities(args interface{}, resp *[]string) error {
	var err error
	*resp, err = s.Impl.GetCapabilities()
	return err
}

// NodeExecutorPlugin est l'implémentation de plugin.Plugin pour notre NodeExecutor.
type NodeExecutorPlugin struct {
	Impl NodeExecutor
}

func (p *NodeExecutorPlugin) Server(b *plugin.MuxBroker) (interface{}, error) {
	return &NodeExecutorRPCServer{Impl: p.Impl}, nil
}

func (p *NodeExecutorPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &NodeExecutorRPC{client: c}, nil
}

// --- Implémentation du wrapper gRPC (plus moderne) ---
// Cette partie est laissée pour référence mais nous utilisons NetRPC pour la simplicité de l'encodage gob.
// Pour gRPC, il faudrait générer le code à partir d'un fichier .proto.

// func (p *NodeExecutorPlugin) GRPCServer(b *plugin.GRPCBroker, s *grpc.Server) error {
// 	return nil // Pas implémenté pour l'instant
// }

// func (p *NodeExecutorPlugin) GRPCClient(ctx context.Context, b *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
// 	return nil, nil // Pas implémenté pour l'instant
// }
