from setuptools import setup, find_packages

with open("README.md", "r", encoding="utf-8") as fh:
    long_description = fh.read()

setup(
    name="xyphos_client",
    version="1.0.0",
    author="Xyphos Team",
    author_email="support@xyphos.io",
    description="Python client for Xyphos KMS - an open-source key management system",
    long_description=long_description,
    long_description_content_type="text/markdown",
    url="https://github.com/theboringhumane/xyphos",
    packages=find_packages(),
    classifiers=[
        "Development Status :: 5 - Production/Stable",
        "Intended Audience :: Developers",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
        "Programming Language :: Python :: 3",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Topic :: Security :: Cryptography",
    ],
    python_requires=">=3.8",
    install_requires=[
        "httpx>=0.24.0",
        "cryptography>=41.0.0",
        "pydantic>=2.0.0",
        "python-dateutil>=2.8.2",
        "requests>=2.28.0",
        "urllib3>=1.26.16",
    ],
    extras_require={
        "dev": [
            "pytest>=7.0.0",
            "pytest-asyncio>=0.21.0",
            "pytest-cov>=4.0.0",
            "black>=23.0.0",
            "isort>=5.12.0",
            "mypy>=1.0.0",
            "types-python-dateutil>=2.8.19",
        ],
    },
    project_urls={
        "Bug Tracker": "https://github.com/theboringhumane/xyphos-client/issues",
        "Documentation": "https://docs.xyphos.io",
        "Source Code": "https://github.com/theboringhumane/xyphos-client",
    },
) 