package k8s

import (
	"admission-webhook-pod/options"
	"bytes"
	"context"
	"fmt"
	"net"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	mutatingWebhookName              = "admission-webhook-pod.zeratullich.com"
	mutationWebhookConfigurationName = "admission-webhook-pod"
)

var webhookLabel = map[string]string{"app": mutationWebhookConfigurationName}

type K8S struct {
	kubernetesClient           *kubernetes.Clientset
	config                     *rest.Config
	parameters                 *options.OptionsParams
	CaPEM, CertPEM, CertKeyPEM *bytes.Buffer
}

func (k *K8S) Run() error {

	serviceName := k.parameters.Service
	namespace := k.parameters.Namespace

	dnsNames := []string{serviceName, fmt.Sprintf("%s.%s", serviceName, namespace), fmt.Sprintf("%s.%s.svc", serviceName, namespace)}
	orgs := mutatingWebhookName
	commonName := fmt.Sprintf("%s.%s.svc", serviceName, namespace)
	var ips []net.IP

	if net.ParseIP(k.parameters.URL) != nil {
		ips = append(ips, net.ParseIP(k.parameters.URL))
	}

	var err error
	k.CaPEM, k.CertPEM, k.CertKeyPEM, err = generateCert([]string{orgs}, dnsNames, commonName, ips)
	if err != nil {
		log.Fatalf("Failed to generate ca and certificate key pair: %v", err)
	}

	var (
		path    = options.MutatePath
		url     string
		service *admissionv1.ServiceReference
	)

	log.Debugf("DEBUG model: %t", k.parameters.IsDebug)
	if k.parameters.IsDebug {
		url = fmt.Sprintf("https://%s:%d%s", k.parameters.URL, k.parameters.Port, path)
		err = k.createMutationWebhook(k.CaPEM, mutationWebhookConfigurationName, mutatingWebhookName, nil, &url)
	} else {
		service = &admissionv1.ServiceReference{
			Name:      k.parameters.Service,
			Namespace: k.parameters.Namespace,
			Path:      &path,
			Port:      &k.parameters.Port,
		}
		logMU, _ := yaml.Marshal(service)
		log.Debugf(string(logMU))
		err = k.createMutationWebhook(k.CaPEM, mutationWebhookConfigurationName, mutatingWebhookName, service, nil)
	}
	return err
}

func (k *K8S) createMutationWebhook(caPEM *bytes.Buffer, mutatingWebhookConfigurationName, webhookName string, service *admissionv1.ServiceReference, url *string) error {

	sideEffect := admissionv1.SideEffectClass(admissionv1.SideEffectClassNone)
	fail := admissionv1.Fail
	mutationWebhook := &admissionv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:   mutatingWebhookConfigurationName,
			Labels: webhookLabel,
		},
		Webhooks: []admissionv1.MutatingWebhook{
			{
				Name: webhookName,
				ClientConfig: admissionv1.WebhookClientConfig{
					Service:  service,
					URL:      url,
					CABundle: caPEM.Bytes(),
				},
				AdmissionReviewVersions: []string{"v1"},
				SideEffects:             &sideEffect,
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
						},
						Rule: admissionv1.Rule{
							APIGroups:   []string{""},
							APIVersions: []string{"v1"},
							Resources:   []string{"pods"},
						},
					},
				},
				NamespaceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{mutationWebhookConfigurationName: "enabled"},
				},
				FailurePolicy: &fail,
			},
		},
	}
	_, err := k.kubernetesClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.Background(), mutationWebhook, metav1.CreateOptions{})
	if err != nil {
		if errors.IsAlreadyExists(err) {
			log.Warnf("MutatingWebhookConfigurations %s is already exist , and deleting it...", mutatingWebhookConfigurationName)
			if err = k.kubernetesClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(context.Background(), mutationWebhookConfigurationName, metav1.DeleteOptions{}); err != nil {
				log.Errorf("Delete mutatingWebhookConfigurations failed, reason is %v", err)
				return err
			} else {
				if _, err = k.kubernetesClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.Background(), mutationWebhook, metav1.CreateOptions{}); err != nil {
					log.Errorf("Create mutatingWebhookConfigurations failed, reason is %v", err)
					return err
				} else {
					log.Infof("Rebuild MutatingWebhookConfigurations %s successfully!", mutatingWebhookConfigurationName)
				}
			}
		} else {
			return err
		}
	} else {
		log.Infof("Create MutatingWebhookConfigurations %s successfully!", mutatingWebhookConfigurationName)
	}
	return nil
}

func NewK8S(options *options.OptionsParams) *K8S {
	k, err := newKubernetesClient(options)
	if err != nil {
		log.Panic(err)
	}
	k.parameters = options
	return k
}
