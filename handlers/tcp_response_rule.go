package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/haproxytech/client-native"
	"github.com/haproxytech/dataplaneapi/haproxy"
	"github.com/haproxytech/dataplaneapi/misc"
	"github.com/haproxytech/dataplaneapi/operations/tcp_response_rule"
)

//CreateTCPResponseRuleHandlerImpl implementation of the CreateTCPResponseRuleHandler interface using client-native client
type CreateTCPResponseRuleHandlerImpl struct {
	Client      *client_native.HAProxyClient
	ReloadAgent *haproxy.ReloadAgent
}

//DeleteTCPResponseRuleHandlerImpl implementation of the DeleteTCPResponseRuleHandler interface using client-native client
type DeleteTCPResponseRuleHandlerImpl struct {
	Client      *client_native.HAProxyClient
	ReloadAgent *haproxy.ReloadAgent
}

//GetTCPResponseRuleHandlerImpl implementation of the GetTCPResponseRuleHandler interface using client-native client
type GetTCPResponseRuleHandlerImpl struct {
	Client *client_native.HAProxyClient
}

//GetTCPResponseRulesHandlerImpl implementation of the GetTCPResponseRulesHandler interface using client-native client
type GetTCPResponseRulesHandlerImpl struct {
	Client *client_native.HAProxyClient
}

//ReplaceTCPResponseRuleHandlerImpl implementation of the ReplaceTCPResponseRuleHandler interface using client-native client
type ReplaceTCPResponseRuleHandlerImpl struct {
	Client      *client_native.HAProxyClient
	ReloadAgent *haproxy.ReloadAgent
}

//Handle executing the request and returning a response
func (h *CreateTCPResponseRuleHandlerImpl) Handle(params tcp_response_rule.CreateTCPResponseRuleParams, principal interface{}) middleware.Responder {
	t := ""
	v := int64(0)
	if params.TransactionID != nil {
		t = *params.TransactionID
	}
	if params.Version != nil {
		v = *params.Version
	}

	err := h.Client.Configuration.CreateTCPResponseRule(params.Backend, params.Data, t, v)
	if err != nil {
		e := misc.HandleError(err)
		return tcp_response_rule.NewCreateTCPResponseRuleDefault(int(*e.Code)).WithPayload(e)
	}
	h.ReloadAgent.Reload()
	return tcp_response_rule.NewCreateTCPResponseRuleCreated().WithPayload(params.Data)
}

//Handle executing the request and returning a response
func (h *DeleteTCPResponseRuleHandlerImpl) Handle(params tcp_response_rule.DeleteTCPResponseRuleParams, principal interface{}) middleware.Responder {
	t := ""
	v := int64(0)
	if params.TransactionID != nil {
		t = *params.TransactionID
	}
	if params.Version != nil {
		v = *params.Version
	}

	err := h.Client.Configuration.DeleteTCPResponseRule(params.ID, params.Backend, t, v)
	if err != nil {
		e := misc.HandleError(err)
		return tcp_response_rule.NewDeleteTCPResponseRuleDefault(int(*e.Code)).WithPayload(e)
	}
	h.ReloadAgent.Reload()
	return tcp_response_rule.NewDeleteTCPResponseRuleNoContent()
}

//Handle executing the request and returning a response
func (h *GetTCPResponseRuleHandlerImpl) Handle(params tcp_response_rule.GetTCPResponseRuleParams, principal interface{}) middleware.Responder {
	t := ""
	if params.TransactionID != nil {
		t = *params.TransactionID
	}

	rule, err := h.Client.Configuration.GetTCPResponseRule(params.ID, params.Backend, t)
	if err != nil {
		e := misc.HandleError(err)
		return tcp_response_rule.NewGetTCPResponseRuleDefault(int(*e.Code)).WithPayload(e)
	}
	return tcp_response_rule.NewGetTCPResponseRuleOK().WithPayload(rule)
}

//Handle executing the request and returning a response
func (h *GetTCPResponseRulesHandlerImpl) Handle(params tcp_response_rule.GetTCPResponseRulesParams, principal interface{}) middleware.Responder {
	t := ""
	if params.TransactionID != nil {
		t = *params.TransactionID
	}

	rules, err := h.Client.Configuration.GetTCPResponseRules(params.Backend, t)
	if err != nil {
		e := misc.HandleError(err)
		return tcp_response_rule.NewGetTCPResponseRulesDefault(int(*e.Code)).WithPayload(e)
	}
	return tcp_response_rule.NewGetTCPResponseRulesOK().WithPayload(rules)
}

//Handle executing the request and returning a response
func (h *ReplaceTCPResponseRuleHandlerImpl) Handle(params tcp_response_rule.ReplaceTCPResponseRuleParams, principal interface{}) middleware.Responder {
	t := ""
	v := int64(0)
	if params.TransactionID != nil {
		t = *params.TransactionID
	}
	if params.Version != nil {
		v = *params.Version
	}

	err := h.Client.Configuration.EditTCPResponseRule(params.ID, params.Backend, params.Data, t, v)
	if err != nil {
		e := misc.HandleError(err)
		return tcp_response_rule.NewReplaceTCPResponseRuleDefault(int(*e.Code)).WithPayload(e)
	}
	h.ReloadAgent.Reload()
	return tcp_response_rule.NewReplaceTCPResponseRuleOK().WithPayload(params.Data)
}