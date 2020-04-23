import React from 'react';

// based on https://rastating.github.io/creating-a-conditional-react-hook/

type useAPIParams = {
  endpoint: string;
  method: string;
  skip: boolean;
  withReCAPTCHA: boolean;
};

export default function useAPI({ endpoint, method, skip, withReCAPTCHA = false }: useAPIParams) {
  const [result, setResult] = React.useState<any>(null);
  const [loading, setLoading] = React.useState(false);
  const [hasError, setHasError] = React.useState(false);
  if (!skip && !loading) {
    setLoading(true);
  }

  React.useEffect(() => {
    if (!skip) {
      const getReCAPTCHAToken = async () => {
        return new Promise((r) => {
          grecaptcha?.ready(() => {
            grecaptcha
              ?.execute('6LcEX-0UAAAAAPakStenDryOkvRgineD9Sn5Xbqg', { action: 'validate_captcha' })
              .then((token: string) => {
                r(token);
              });
          });
        });
      };
      const executeRequest = async () => {
        try {
          let additionalParams = '';
          if (withReCAPTCHA) {
            const token = await getReCAPTCHAToken();
            additionalParams = `&recaptcha=${token}`;
          }
          const res = await fetch(`${endpoint}${additionalParams}`, {
            method,
          });
          setResult(await res.json());
        } catch (error) {
          setHasError(true);
        }
        setLoading(false);
      };
      executeRequest();
    }
  }, [skip, endpoint, method, withReCAPTCHA]);

  return [loading, hasError, result];
}
